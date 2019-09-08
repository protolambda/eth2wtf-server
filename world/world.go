package world

import (
	"bytes"
	"encoding/binary"
	"eth2wtf-server/common"
	"log"
	"time"
)

const EV_LOG_SIZE = 1000

type World struct {
	Input          chan common.Event
	Logger         *log.Logger
	eventLog       [EV_LOG_SIZE]common.Event
	NextEventIndex uint64
}

func NewWorld() *World {
	return &World{
	}
}

func (w *World) HandleRequest(v common.Viewer, start uint64, end uint64) {
	min := v.EventIndex()
	if start > min {
		min = start
	}
	if min+EV_LOG_SIZE < w.NextEventIndex {
		min = w.NextEventIndex
	}
	// TODO: could be a pooled buffer
	var buf bytes.Buffer
	{ // encode the offset we start at.
		lenVal := [8]byte{}
		dat := buf.Bytes()
		binary.LittleEndian.PutUint64(lenVal[:], min)
		buf.Write(dat[:])
	}
	for i := min; i < end; i++ {
		ev := w.eventLog[i%EV_LOG_SIZE]
		if ev == nil {
			w.Logger.Printf("cannot find event at index %d, stopping iteration", i)
			break
		}
		if err := ev.Serialize(&buf); err != nil {
			w.Logger.Printf("failed to serialize event %d: %v", i, err)
			return
		}
	}
	v.Send(buf.Bytes())
}

func (w *World) HeartBeat(getViewers common.ViewersGetter) {
	for {
		for _, v := range getViewers() {
			w.HandleRequest(v, 0, uint64(len(w.eventLog)))
		}
		time.Sleep(time.Second)
	}
}

func (w *World) Process() {
	for {
		ev, ok := <-w.Input
		w.eventLog[w.NextEventIndex] = ev
		w.NextEventIndex++
		if !ok {
			return
		}
	}
}
