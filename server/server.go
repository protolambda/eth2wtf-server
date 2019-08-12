package server

import (
	"encoding/binary"
	"eth2wtf-server/client"
	"eth2wtf-server/hub"
	"net/http"
)

type Server struct {
	clientHub *hub.Hub
	chunks map[ChunkID]Chunk
}

func NewServer() *Server {
	return &Server{
		clientHub: hub.NewHub(),
		chunks: make(map[ChunkID]Chunk),
	}
}

func (s *Server) Run() {
	s.clientHub.Run()
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	s.clientHub.ServeWs(w, r, s.NewClientHandler)
}

func (s *Server) NewClientHandler(send chan<- []byte) client.ClientHandler {
	return NewClientHandler(s.GetChunk, send)
}

func (s *Server) GetChunk(ctID ContentID, chunkID ChunkID) ChunkContentHandler {
	if c, ok := s.chunks[chunkID]; ok {
		return c.GetContentChunk(ctID)
	} else {
		return nil
	}
}

type ContentChunk struct {
	Content ChunkContentHandler
	CtID ContentID
	ChunkID ChunkID
}

func (c *ContentChunk) HandleRequest(msg []byte, send func(res []byte)) {
	c.Content.HandleRequest(msg, func(res []byte) {
		// Add the message type, content type, chunk ID to the message as prefix.
		dataLen := 2 + 4 + len(res)
		data := make([]byte, dataLen, dataLen)
		data[0] = 1
		data[1] = byte(c.CtID)
		binary.LittleEndian.PutUint32(data[2:6], uint32(c.ChunkID))
		copy(data[6:], res)
		send(data)
	})
}

type Chunk struct {
	Headers ContentChunk
}

func (c *Chunk) GetContentChunk(ctID ContentID) *ContentChunk {
	switch ctID {
	case 1:
		return &c.Headers
	default:
		return nil
	}
}