package common

type ChunkID uint32

type ContentID byte

type ChunkContentHandler interface {
	HandleRequest(msg []byte, send func(res []byte))
}

type Viewport struct {
	Min ChunkID
	Max ChunkID
}
