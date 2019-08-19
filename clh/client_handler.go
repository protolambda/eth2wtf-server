package clh

import (
	"encoding/binary"
	. "eth2wtf-server/common"
	"fmt"
)

type WorldLike interface {
	GetChunk(ctID ContentID, chunkID ChunkID) ChunkContentHandler
}

type ClientHandler struct {
	world    WorldLike
	send     chan<- []byte
	viewport Viewport
}

func NewClientHandler(world WorldLike, send chan<- []byte) *ClientHandler {
	return &ClientHandler{
		world: world,
		send:  send,
	}
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
	switch msgType {
	// 1: messages that are content-specific updates.
	case 1:
		if len(msg) < 6 {
			return
		}
		ctID := ContentID(msg[1])
		chunkID := ChunkID(binary.LittleEndian.Uint32(msg[2:6]))
		chunk := ch.world.GetChunk(ctID, chunkID)
		if chunk == nil {
			fmt.Printf("ignoring request for unknown chunk content: type: %d chunk: %d\n", ctID, chunkID)
		} else {
			chunk.HandleRequest(msg[6:], ch.Send)
		}
	case 2:
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
