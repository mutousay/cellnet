package socket

import (
	"bytes"
	"encoding/binary"
	"github.com/mutousay/cellnet"
	"io"
	"sync"
)

type PrivatePacketReader struct {
	recvser uint16
}

func (self *PrivatePacketReader) Call(ev *cellnet.Event) {

	headReader := bytes.NewReader(ev.Data)
	//log.Debugln("PrivatePacketReader  data", headReader)
	// 读取序号
	var ser uint16
	if err := binary.Read(headReader, binary.LittleEndian, &ser); err != nil {
		log.Debugln("PrivatePacketReader  1111")
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	// 读取ID
	if err := binary.Read(headReader, binary.LittleEndian, &ev.MsgID); err != nil {
		log.Debugln("PrivatePacketReader  2222")
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}
	//log.Debugln("PrivatePacketReader  msgID", ev.MsgID)

	// 读取Payload大小
	var bodySize uint32
	if err := binary.Read(headReader, binary.LittleEndian, &bodySize); err != nil {
		log.Debugln("PrivatePacketReader  3333")
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}
	//log.Debugln("PrivatePacketReader  size", bodySize)

	maxPacketSize := ev.Ses.FromPeer().(SocketOptions).MaxPacketSize()
	// 封包太大
	if maxPacketSize > 0 && int(bodySize) > maxPacketSize {
		log.Debugln("PrivatePacketReader  44444")
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	// 序列号不匹配  TODO 如果客户端和游戏服同时给网关发消息，则会导致这边的recvser 会不同步增加
	// 这边需要等调整网关设计之后,再开起来，对外的网关和对内的网关都区分开就不会导致这个问题
	//if self.recvser != ser{
	// if self.recvser != ser && ser != 0 { //客户端第一次发的是0
	// 	log.Debugln("PrivatePacketReader  5555")
	// 	log.Debugln("PrivatePacketReader recvser ", self.recvser)
	// 	log.Debugln("PrivatePacketReader ser ", ser)
	// 	ev.SetResult(cellnet.Result_PackageCrack)
	// 	return
	// }

	reader := ev.Ses.(interface {
		DataSource() io.ReadWriter
	}).DataSource()

	// 读取数据
	dataBuffer := make([]byte, bodySize)
	if _, err := io.ReadFull(reader, dataBuffer); err != nil {
		log.Debugln("PrivatePacketReader  66666")
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	//log.Debugln("PrivatePacketReader  data bufffer", dataBuffer)

	ev.Data = dataBuffer

	// 增加序列号值
	self.recvser++
}

//TODO reader 问题 客户端处理链到PrivatePacketReader 之后停止
func NewPrivatePacketReader() cellnet.EventHandler {
	return &PrivatePacketReader{
		recvser: 1,
	}
}

type PrivatePacketWriter struct {
	sendser      uint16
	sendtagGuard sync.RWMutex
}

func (self *PrivatePacketWriter) Call(ev *cellnet.Event) {

	// 防止将Send放在go内造成的多线程冲突问题
	self.sendtagGuard.Lock()
	defer self.sendtagGuard.Unlock()

	var outputHeadBuffer bytes.Buffer

	// 写序号
	//log.Debugln("PrivatePacketWriter sendser ", self.sendser)
	if err := binary.Write(&outputHeadBuffer, binary.LittleEndian, self.sendser); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	// 写ID
	if err := binary.Write(&outputHeadBuffer, binary.LittleEndian, ev.MsgID); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	// 写包大小
	if err := binary.Write(&outputHeadBuffer, binary.LittleEndian, uint32(len(ev.Data))); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	binary.Write(&outputHeadBuffer, binary.LittleEndian, ev.Data)

	// 增加序号值
	self.sendser++

	ev.Data = outputHeadBuffer.Bytes()
}

func NewPrivatePacketWriter() cellnet.EventHandler {
	return &PrivatePacketWriter{
		sendser: 1,
	}
}
