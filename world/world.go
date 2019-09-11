package world

import (
	"bytes"
	"encoding/binary"
	"eth2wtf-server/common"
	"log"
	"time"
)

const EV_LOG_SIZE = 1000

const MAX_EVS_PER_RESPONSE = 20


type World struct {
	Input          chan common.Event
	Logger         *log.Logger
	eventLog       [EV_LOG_SIZE]common.Event
	NextEventIndex common.EventIndex
}

func NewWorld(l *log.Logger) *World {
	return &World{
		Input: make(chan common.Event),
		Logger: l,
	}
}

func (w *World) InputEv(ev common.Event) bool {
	w.Input <- ev
	// TODO return false when already closed
	return true
}

func (w *World) HandleRequest(v common.Viewer, start common.EventIndex, end common.EventIndex) {
	min := v.EventIndex()
	if start > min {
		min = start
	}
	if min+EV_LOG_SIZE < w.NextEventIndex {
		min = w.NextEventIndex - EV_LOG_SIZE
	}
	max := min + MAX_EVS_PER_RESPONSE
	if max > end {
		max = end
	}
	if max > w.NextEventIndex {
		max = w.NextEventIndex
	}
	w.Logger.Printf("serving %d to %d (%d, %d)\n", min, max, start, end)
	// TODO: could be a pooled buffer
	var buf bytes.Buffer
	{
		// Write the message response type
		buf.WriteByte(1)
	}
	{ // encode the event index we start at.
		indexV := [4]byte{}
		w.Logger.Printf("responding index: %v\n", min)
		binary.LittleEndian.PutUint32(indexV[:], uint32(min))
		buf.Write(indexV[:])
	}
	var evMsgBuf bytes.Buffer
	for i := min; i < max; i++ {
		ev := w.eventLog[i%EV_LOG_SIZE]
		if ev == nil {
			w.Logger.Printf("cannot find event at index %d, stopping iteration", i)
			break
		}
		if err := buf.WriteByte(byte(ev.EventType())); err != nil {
			w.Logger.Printf("failed to write event type %d: %v", i, err)
			return
		}
		if err := ev.Serialize(&evMsgBuf); err != nil {
			w.Logger.Printf("failed to serialize event %d: %v", i, err)
			return
		}
		evByteLen := [4]byte{}
		dat := evMsgBuf.Bytes()
		binary.LittleEndian.PutUint32(evByteLen[:], uint32(len(dat)))
		buf.Write(evByteLen[:])
		buf.Write(dat)
		evMsgBuf.Reset()
	}
	outDat := buf.Bytes()
	w.Logger.Printf("out: %x\n", outDat)
	v.Send(outDat)
}

func (w *World) HeartBeat(getViewers common.ViewersGetter) {
	for {
		w.Logger.Printf("heartbeat, ev index: %d\n", w.NextEventIndex)
		for _, v := range getViewers() {
			w.HandleRequest(v, 0, w.NextEventIndex)
		}
		time.Sleep(time.Second)
	}
}

func (w *World) Process() {
	w.Logger.Println("started event processing")
	for {
		ev, ok := <-w.Input
		w.Logger.Printf("processing ev: %v\n", w.NextEventIndex)
		w.eventLog[w.NextEventIndex % EV_LOG_SIZE] = ev
		w.NextEventIndex++
		if !ok {
			w.Logger.Println("stopping processing")
			return
		}
	}
}
