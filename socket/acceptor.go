package socket

import (
	"net"
	"time"
	"github.com/mutousay/cellnet"
	"github.com/mutousay/cellnet/extend"

	kcp "github.com/xtaci/kcp-go"
)

type socketAcceptor struct {
	*socketPeer

	//listener net.Listener
	listener *kcp.Listener
}

func (self *socketAcceptor) Start(address string) cellnet.Peer {

	self.waitStopFinished()

	if self.IsRunning() {
		return self
	}

	self.SetAddress(address)

	//kcp修改
	//ln, err := net.Listen("tcp", address)
	ln, err := kcp.Listen(address)
	if err != nil {
		log.Errorf("#listen failed(%s) %v", self.NameOrAddress(), err.Error())
		return self		
	}
	kcpListener := ln.(*kcp.Listener)
	kcpListener.SetReadBuffer(4 * 1024 * 1024)
	kcpListener.SetWriteBuffer(4 * 1024 * 1024)
	kcpListener.SetDSCP(46)
	self.listener = kcpListener

	log.Infof("#listen(%s) %s", self.Name(), self.Address())

	// 接受线程
	go self.accept()

	return self
}

func (self *socketAcceptor) accept() {

	self.SetRunning(true)

	for {
		log.Infoln("wait next connection ")
		conn, err := self.listener.Accept()
		//kcp修改
		conn.(*kcp.UDPSession).SetNoDelay(1, 30, 2, 1)
		if self.isStopping() {
			break
		}

		if err != nil {

			// 调试状态时, 才打出accept的具体错误
			if log.IsDebugEnabled() {
				log.Errorf("#accept failed(%s) %v", self.NameOrAddress(), err.Error())
			}

			extend.PostSystemEvent(nil, cellnet.Event_AcceptFailed, self.ChainListRecv(), errToResult(err))

			break
		}

		// 处理连接进入独立线程, 防止accept无法响应
		go self.onAccepted(conn)

	}

	self.SetRunning(false)

	self.endStopping()
}

func (self *socketAcceptor) onAccepted(conn net.Conn) {

	ses := newSession(conn, self)
	conn.(*kcp.UDPSession).SetStreamMode(true)
	conn.(*kcp.UDPSession).SetWindowSize(4096, 4096)
	conn.(*kcp.UDPSession).SetNoDelay(1, 10, 2, 1)
	conn.(*kcp.UDPSession).SetDSCP(46)
	conn.(*kcp.UDPSession).SetMtu(1400)
	conn.(*kcp.UDPSession).SetACKNoDelay(false)
	conn.(*kcp.UDPSession).SetReadDeadline(time.Now().Add(time.Hour))
	conn.(*kcp.UDPSession).SetWriteDeadline(time.Now().Add(time.Hour))
	
	// 添加到管理器
	self.Add(ses)
	log.Debugf("onAccepted session sid:%d", ses.ID)
	// 断开后从管理器移除
	ses.OnClose = func() {
		self.Remove(ses)
	}

	ses.run()

	// 通知逻辑
	extend.PostSystemEvent(ses, cellnet.Event_Accepted, self.ChainListRecv(), cellnet.Result_OK)
}

func (self *socketAcceptor) Stop() {

	if !self.IsRunning() {
		return
	}

	if self.isStopping() {
		return
	}

	self.startStopping()

	self.listener.Close()

	// 断开所有连接
	self.CloseAllSession()

	// 等待线程结束
	self.waitStopFinished()
}

func NewAcceptor(q cellnet.EventQueue) cellnet.Peer {

	self := &socketAcceptor{
		socketPeer: newSocketPeer(q, cellnet.NewSessionManager()),
	}

	return self
}
