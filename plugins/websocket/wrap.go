package websocket

import (
	"log"

	"github.com/Allenxuxu/gev/connection"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/Allenxuxu/gev/plugins/websocket/ws/util"
)

// WSHandler WebSocket Server 注册接口
type WSHandler interface {
	OnConnect(c *connection.Connection)
	OnMessage(c *connection.Connection, msg []byte) (ws.MessageType, []byte)
	OnClose(c *connection.Connection)
}

type handlerWrap struct {
	wsHandler WSHandler
	Upgrade   *ws.Upgrader
}

func NewHandlerWrap(u *ws.Upgrader, wsHandler WSHandler) *handlerWrap {
	return &handlerWrap{
		wsHandler: wsHandler,
		Upgrade:   u,
	}
}

func (s *handlerWrap) OnConnect(c *connection.Connection) {
	s.wsHandler.OnConnect(c)
}

func (s *handlerWrap) OnMessage(c *connection.Connection, ctx interface{}, payload []byte) []byte {
	header, ok := ctx.(*ws.Header)
	if !ok && len(payload) != 0 { // 升级协议 握手
		return payload
	}

	if ok {
		if header.OpCode.IsControl() {
			var (
				out []byte
				err error
			)
			switch header.OpCode {
			case ws.OpClose:
				out, err = util.HandleClose(header, payload)
				if err != nil {
					log.Println(err)
				}
				_ = c.ShutdownWrite()
			case ws.OpPing:
				out, err = util.HandlePing(payload)
				if err != nil {
					log.Println(err)
				}
			case ws.OpPong:
				out, err = util.HandlePong(payload)
				if err != nil {
					log.Println(err)
				}
			}
			return out
		}

		messageType, out := s.wsHandler.OnMessage(c, payload)
		if len(out) > 0 {
			var frame *ws.Frame
			switch messageType {
			case ws.MessageBinary:
				frame = ws.NewBinaryFrame(out)
			case ws.MessageText:
				frame = ws.NewTextFrame(out)
			}
			var err error
			out, err = ws.FrameToBytes(frame)
			if err != nil {
				log.Println(err)
			}

			return out
		}
	}
	return nil
}

func (s *handlerWrap) OnClose(c *connection.Connection) {
	s.wsHandler.OnClose(c)
}
