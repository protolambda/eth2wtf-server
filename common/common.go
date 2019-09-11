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
	HandleRequest(v Viewer, start EventIndex, end EventIndex)
}

type EventType byte

const (
	HeaderEventID EventType = 1 + iota
)

type EventIndex uint32

type Event interface {
	EventType() EventType
	Serialize(w io.Writer) error
}

type Viewer interface {
	EventIndex() EventIndex
	ReceivePort
}

type ViewersGetter func() []Viewer
