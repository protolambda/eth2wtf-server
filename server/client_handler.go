package server

import (
	"encoding/binary"
	"fmt"
)

type ChunkGetter func(ctID ContentID, chunkID ChunkID) ChunkContentHandler

type ClientHandler struct {
	getChunk ChunkGetter
	send chan<- []byte
}

func NewClientHandler(getChunk ChunkGetter, send chan<- []byte) *ClientHandler {
	return &ClientHandler{
		getChunk: getChunk,
		send:     send,
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
		if len(msg) < 2 {
			return
		}
		ctID := ContentID(msg[1])
		chunkID := ChunkID(binary.LittleEndian.Uint32(msg[2:6]))
		chunk := ch.getChunk(ctID, chunkID)
		if chunk == nil {
			fmt.Printf("ignoring request for unknown chunk content: type: %d chunk: %d\n", ctID, chunkID)
		} else {
			chunk.HandleRequest(msg[6:], ch.Send)
		}
	default:
		fmt.Printf("unrecognized msgType: %d\n", msgType)
	}
}
