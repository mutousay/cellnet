package main

import (
	"github.com/davyxu/golog"
	"github.com/mutousay/cellnet"
	"github.com/mutousay/cellnet/examples/chat/proto/chatproto"
	"github.com/mutousay/cellnet/socket"
)

var log *golog.Logger = golog.New("main")

func main() {
	queue := cellnet.NewEventQueue()

	peer := socket.NewAcceptor(queue).Start("127.0.0.1:8801")
	peer.SetName("client")

	cellnet.RegisterMessage(peer, "chatproto.ChatREQ", func(ev *cellnet.Event) {
		msg := ev.Msg.(*chatproto.ChatREQ)

		ack := chatproto.ChatACK{
			Id:      ev.Ses.ID(),
			Content: msg.Content,
		}

		// 广播给所有连接
		peer.VisitSession(func(ses cellnet.Session) bool {

			ses.Send(&ack)

			return true
		})

	})

	queue.StartLoop()

	queue.Wait()

	peer.Stop()
}
