package world

import (
	"encoding/binary"
	. "eth2wtf-server/common"
	. "eth2wtf-server/common/contenttyp"
	"eth2wtf-server/world/headers"
	"log"
)

type World struct {
	chunks map[ChunkID]*Chunk
}

func NewWorld() *World {
	return &World{
		chunks: make(map[ChunkID]*Chunk),
	}
}

func (s *World) GetChunk(chunkID ChunkID) ChunkHandler {
	if c, ok := s.chunks[chunkID]; ok {
		return c
	} else {
		return nil
	}
}

// GetChunk, but creates the chunk if it doesn't already exist
func (s *World) CreateChunkMaybe(chunkID ChunkID) ChunkHandler {
	if v, ok := s.chunks[chunkID]; ok {
		return v
	} else {
		v = &Chunk{chunkID: chunkID}
		s.chunks[chunkID] = v
		return v
	}
}

type Chunk struct {
	headers headers.HeadersChunk
	chunkID ChunkID
}

func (c *Chunk) HeadersChunk() *headers.HeadersChunk {
	return &c.headers
}

func (c *Chunk) HandleRequest(logger *log.Logger, port ReceivePort, ctID ContentID,  msg []byte) {
	content := c.GetContent(ctID)
	if content == nil {
		logger.Printf("Content type %d was not recognized for chunk %d", ctID, c.chunkID)
		return
	}
	content.HandleRequest(logger, ReceivePortFn(func(res []byte) {
		// Add the message type, content type, chunk ID to the message as prefix.
		dataLen := 2 + 4 + len(res)
		data := make([]byte, dataLen, dataLen)
		data[0] = 1
		data[1] = byte(ctID)
		binary.LittleEndian.PutUint32(data[2:6], uint32(c.chunkID))
		copy(data[6:], res)
		port.Send(data)
	}), msg)
}

func (c *Chunk) GetContent(ctID ContentID) ContentHandler {
	switch ctID {
	case 1:
		return &c.headers
	default:
		return nil
	}
}
