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

func (s *World) ChunkGetter(ctID ContentID) ChunkGetter {
	return func(id ChunkID) ChunkContent {
		s.CreateChunkMaybe(id)
		return s.GetChunk(ctID, id)
	}
}

func (s *World) GetChunk(ctID ContentID, chunkID ChunkID) ChunkContent {
	if c, ok := s.chunks[chunkID]; ok {
		return &ChunkView{Content: c.GetContentChunk(ctID), CtID: ctID, ChunkID: chunkID}
	} else {
		return nil
	}
}

func (s *World) CreateChunkMaybe(chunkID ChunkID) {
	if _, ok := s.chunks[chunkID]; !ok {
		s.chunks[chunkID] = new(Chunk)
	}
}

// wrapper around content to provide content and chunk IDs, wrapping the handle-request responses.
type ChunkView struct {
	Content ChunkContent
	CtID    ContentID
	ChunkID ChunkID
}

func (c *ChunkView) HandleRequest(logger *log.Logger, port ReceivePort, msg []byte) {
	c.Content.HandleRequest(logger, ReceivePortFn(func(res []byte) {
		// Add the message type, content type, chunk ID to the message as prefix.
		dataLen := 2 + 4 + len(res)
		data := make([]byte, dataLen, dataLen)
		data[0] = 1
		data[1] = byte(c.CtID)
		binary.LittleEndian.PutUint32(data[2:6], uint32(c.ChunkID))
		copy(data[6:], res)
		port.Send(data)
	}), msg)
}

type Chunk struct {
	Headers headers.HeadersChunk
}

func (c *Chunk) GetContentChunk(ctID ContentID) ChunkContent {
	switch ctID {
	case 1:
		return &c.Headers
	default:
		return nil
	}
}
