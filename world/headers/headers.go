package headers

import (
	"bytes"
	"encoding/binary"
	"eth2wtf-server/common"
	"fmt"
	bh "github.com/protolambda/zrnt/eth2/beacon/header"
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"io"
	"log"
	"math/rand"
	"time"
)

type HeaderData struct {
	Header *bh.BeaconBlockHeader
	Root core.Root
}

var hdrDatSSZ = zssz.GetSSZ((*HeaderData)(nil))

type HeadersProducer struct {
	Headers chan *HeaderData
	Logger *log.Logger
	Closed bool
}

func simHeader(parent *HeaderData, seed uint64) *HeaderData {
	// TODO: customize genesis?
	var h *bh.BeaconBlockHeader
	if parent == nil || parent.Header == nil {
		h = new(bh.BeaconBlockHeader)
	} else {
		h = &bh.BeaconBlockHeader{
			Slot:       parent.Header.Slot + 1,
			ParentRoot: parent.Root,
			StateRoot:  core.Root{},
			BodyRoot:   core.Root{42},
			Signature:  core.BLSSignature{1, 2, 3, 4},
		}
	}
	binary.LittleEndian.PutUint64(h.StateRoot[:], seed)
	return &HeaderData{
		Header: h,
		Root:   ssz.HashTreeRoot(h, bh.BeaconBlockHeaderSSZ),
	}
}

func (hp *HeadersProducer) Mock() {
	lookback := 100
	lastX := make([]*HeaderData, 0, lookback)
	i := 0
	for {
		if hp.Closed {
			return
		}
		pi := rand.Intn(lookback)

		var parent *HeaderData
		if len(lastX) > pi {
			// may be nil
			parent = lastX[len(lastX) - 1 - pi]
		} else {
			parent = nil
		}
		newHeader := simHeader(parent, uint64(i))
		fmt.Printf("header: %d %x -> %x  (%x)\n", newHeader.Header.Slot, newHeader.Header.ParentRoot, newHeader.Root, newHeader.Header.StateRoot)
		lastX = append(lastX, newHeader)
		if len(lastX) > 9 {
			copy(lastX, lastX[1:])
			lastX = lastX[:len(lastX) - 1]
		}
		i++
		hp.Headers <- newHeader
		time.Sleep(200 * time.Millisecond)
	}
}

func (hp *HeadersProducer) Close() {
	hp.Closed = true
	close(hp.Headers)
}

type HeaderEvent struct {
	hd *HeaderData
}

func (he *HeaderEvent) Serialize(w io.Writer) error {
	// TODO could be pooled
	var buf bytes.Buffer
	err := zssz.Encode(&buf, he.hd, hdrDatSSZ)
	if err != nil {
		return err
	}
	lenVal := [4]byte{}
	dat := buf.Bytes()
	binary.LittleEndian.PutUint32(lenVal[:], uint32(len(dat)))
	if _, err := w.Write(lenVal[:]); err != nil {
		return err
	}
	if _, err := w.Write(dat); err != nil {
		return err
	}
	return nil
}

func NewHeaderEvent(hd *HeaderData) *HeaderEvent {
	return &HeaderEvent{
		hd: hd,
	}
}

// The only (synchronous) writer to headers of any chunk
func Process(input chan *HeaderData, output func(ev common.Event) bool) {
	for {
		if h, ok := <-input; ok {
			if !output(NewHeaderEvent(h)) {
				break
			}
		} else {
			break
		}
	}
}
