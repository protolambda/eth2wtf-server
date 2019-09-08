package common

import "io"

type ReceivePort interface{
	Send(msg []byte)
}

type ReceivePortFn func (msg []byte)

func (r ReceivePortFn) Send(msg []byte) {
	r(msg)
}

type RequestHandler interface {
	HandleRequest(v Viewer, start uint64, end uint64)
}

type Event interface {
	Serialize(w io.Writer) error
}

type Viewer interface {
	EventIndex() uint64
	ReceivePort
}

type ViewersGetter func() []Viewer
