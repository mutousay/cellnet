package cellnet

import (
	"io"
	//"runtime/debug"
	"sync"
)

type FixedLengthFrameReader struct {
	headerBuffer []byte
}

func (self *FixedLengthFrameReader) Call(ev *Event) {
	reader := ev.Ses.(interface {
		DataSource() io.ReadWriter
	}).DataSource()
	_, err := io.ReadFull(reader, self.headerBuffer)
	if err != nil {
		log.Debugln("FixedLengthFrameReader err", err)
		ev.SetResult(Result_SocketError)
		//TODO 考虑获取sessionID,如果是服务器的session则不进行关闭，如果是客户端的session则关闭
		return
	}
	ev.Data = self.headerBuffer
}

func NewFixedLengthFrameReader(size int) EventHandler {
	return &FixedLengthFrameReader{
		headerBuffer: make([]byte, size),
	}
}

type FixedLengthFrameWriter struct {
	sendser      uint16
	sendtagGuard sync.RWMutex
}

func (self *FixedLengthFrameWriter) Call(ev *Event) {

	writer := ev.Ses.(interface {
		DataSource() io.ReadWriter
	}).DataSource()

	err := writeFull(writer, ev.Data)

	if err != nil {
		ev.SetResult(Result_PackageCrack)
		return
	}

}

// 完整发送所有封包
func writeFull(writer io.ReadWriter, p []byte) error {

	sizeToWrite := len(p)

	for {

		n, err := writer.Write(p)

		if err != nil {
			return err
		}

		if n >= sizeToWrite {
			break
		}

		copy(p[0:sizeToWrite-n], p[n:sizeToWrite])
		sizeToWrite -= n
	}

	return nil

}

func NewFixedLengthFrameWriter() EventHandler {
	return &FixedLengthFrameWriter{}
}
