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
	eventIndex EventIndex
}

func NewClientHandler(rangeHandler RequestHandler, send chan<- []byte) *ClientHandler {
	return &ClientHandler{
		rangeHandler: rangeHandler,
		send:  send,
	}
}

func (ch *ClientHandler) EventIndex() EventIndex {
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
		if len(msg) != 9 {
			fmt.Printf("msg invalid size: %d\n", len(msg))
			return
		}
		start := EventIndex(binary.LittleEndian.Uint32(msg[1:5]))
		end := EventIndex(binary.LittleEndian.Uint32(msg[5:9]))
		fmt.Printf("client requested range [%d, %d)\n", start, end)
		ch.rangeHandler.HandleRequest(ch, start, end)
	case msgtyp.EventIndexUpdate:
		if len(msg) != 5 {
			fmt.Printf("msg invalid size: %d\n", len(msg))
			return
		}
		ch.eventIndex = EventIndex(binary.LittleEndian.Uint32(msg[1:5]))
		fmt.Printf("client changed event index to %d\n", ch.eventIndex)
	default:
		fmt.Printf("unrecognized msgType: %d\n", msgType)
	}
}
