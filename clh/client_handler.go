package clh

import (
	"encoding/binary"
	. "eth2wtf-server/common"
	"eth2wtf-server/common/contenttyp"
	"eth2wtf-server/common/msgtyp"
	"fmt"
	"log"
)

type WorldLike interface {
	GetChunk(chunkID ChunkID) ChunkHandler
}

type ClientHandler struct {
	world    WorldLike
	logger   *log.Logger
	send     chan<- []byte
	viewport Viewport
}

func NewClientHandler(world WorldLike, send chan<- []byte) *ClientHandler {
	return &ClientHandler{
		world: world,
		send:  send,
	}
}

func (ch *ClientHandler) IsViewing(id ChunkID) bool {
	return id >= ch.viewport.Min && id < ch.viewport.Max
}

func (ch *ClientHandler) Viewport() Viewport {
	return ch.viewport
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
	// 1: messages that are content-specific updates.
	case msgtyp.ChunkRequest:
		if len(msg) < 6 {
			return
		}
		ctID := contenttyp.ContentID(msg[1])
		chunkID := ChunkID(binary.LittleEndian.Uint32(msg[2:6]))
		handler := ch.world.GetChunk(chunkID)
		if handler == nil {
			fmt.Printf("ignoring request for unknown chunk: type: %d chunk: %d\n", ctID, chunkID)
		} else {
			handler.HandleRequest(ch.logger, ch, ctID, msg[6:])
		}
	case msgtyp.ViewportUpdate:
		ch.viewport = Viewport{
			Min: ChunkID(binary.LittleEndian.Uint32(msg[2:6])),
			Max: ChunkID(binary.LittleEndian.Uint32(msg[6:10])),
		}
		// TODO share viewport with world, use pointer ID
		fmt.Printf("client changed viewport to [%d, %d]\n", ch.viewport.Min, ch.viewport.Max)
	default:
		fmt.Printf("unrecognized msgType: %d\n", msgType)
	}
}
