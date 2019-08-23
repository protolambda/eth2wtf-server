package common

import (
	"eth2wtf-server/common/contenttyp"
	"log"
)

type ReceivePort interface{
	Send(msg []byte)
}

type ReceivePortFn func (msg []byte)

func (r ReceivePortFn) Send(msg []byte) {
	r.Send(msg)
}

type ChunkID uint32

type ChunkHandler interface {
	HandleRequest(logger *log.Logger, recv ReceivePort, ctID contenttyp.ContentID,  msg []byte)
}

type ContentHandler interface {
	HandleRequest(logger *log.Logger, recv ReceivePort, msg []byte)
}

type ChunkGetter func(id ChunkID) ChunkHandler

type ViewersGetter func(id ChunkID) []ReceivePort

type Viewport struct {
	// inclusive
	Min ChunkID
	// exclusive
	Max ChunkID
}
