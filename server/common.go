package server

type ChunkID uint32

type ContentID byte

type ChunkContentHandler interface {
	HandleRequest(msg []byte, send func(res []byte))
}
