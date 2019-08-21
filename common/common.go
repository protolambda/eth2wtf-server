package common

import "log"

type ReceivePort interface{
	Send(msg []byte)
}

type ReceivePortFn func (msg []byte)

func (r ReceivePortFn) Send(msg []byte) {
	r.Send(msg)
}

type ChunkID uint32

type ChunkContent interface {
	HandleRequest(logger *log.Logger, recv ReceivePort, msg []byte)
}

type ChunkGetter func(id ChunkID) ChunkContent

type ViewersGetter func(id ChunkID) []ReceivePort

type Viewport struct {
	// inclusive
	Min ChunkID
	// exclusive
	Max ChunkID
}
