package clh

import (
	"encoding/binary"
	. "eth2wtf-server/common"
	"eth2wtf-server/common/msgtyp"
	"fmt"
)

type ClientHandler struct {
	rangeHandler RequestHandler
	send     chan<- []byte
	eventIndex uint64
}

func NewClientHandler(rangeHandler RequestHandler, send chan<- []byte) *ClientHandler {
	return &ClientHandler{
		rangeHandler: rangeHandler,
		send:  send,
	}
}

func (ch *ClientHandler) EventIndex() uint64 {
	return ch.eventIndex
}

func (ch *ClientHandler) Close() {
	fmt.Println("closing client")
	// TODO: unsubscribe from realtime updates
}

func (ch *ClientHandler) Send(msg []byte) {
	ch.send <- msg
}

func (ch *ClientHandler) OnMessage(msg []byte) {
	if len(msg) < 1 {
		return
	}
	msgType := msg[0]
	switch msgtyp.MsgTypeID(msgType) {
	case msgtyp.EventRangeRequest:
		if len(msg) != 17 {
			fmt.Printf("msg invalid size: %d", len(msg))
			return
		}
		start := binary.LittleEndian.Uint64(msg[1:9])
		end := binary.LittleEndian.Uint64(msg[9:17])
		fmt.Printf("client requested range [%d, %d)\n", start, end)
		ch.rangeHandler.HandleRequest(ch, start, end)
	case msgtyp.EventIndexUpdate:
		if len(msg) != 9 {
			fmt.Printf("msg invalid size: %d", len(msg))
			return
		}
		ch.eventIndex = binary.LittleEndian.Uint64(msg[1:9])
		fmt.Printf("client changed event index to %d\n", ch.eventIndex)
	default:
		fmt.Printf("unrecognized msgType: %d\n", msgType)
	}
}
