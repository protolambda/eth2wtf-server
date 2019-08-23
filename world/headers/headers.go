package headers

import (
	"bytes"
	"eth2wtf-server/common"
	. "eth2wtf-server/common"
	bh "github.com/protolambda/zrnt/eth2/beacon/header"
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"log"
	"math/rand"
	"time"
)

const slotsPerChunk = 16

type WithHeaders interface {
	HeadersChunk() *HeadersChunk
}

type HeadersChunk struct {
	Headers []*bh.BeaconBlockHeader
}

func (c *HeadersChunk) HandleRequest(logger *log.Logger, recv ReceivePort, msg []byte) {
	var indices HeaderIndices
	if err := zssz.Decode(bytes.NewReader(msg), uint64(len(msg)), &indices, hdrIndicesSSZ); err != nil {
		logger.Printf("malformed request: %v", err)
	}
	c.PushHeaders(logger, recv, indices...)
}

type HeaderIndices []uint32

func (*HeaderIndices) Limit() uint64 {
	return 1 << 32
}

var hdrIndicesSSZ = zssz.GetSSZ((*HeaderIndices)(nil))

type Headers []*bh.BeaconBlockHeader

func (*Headers) Limit() uint64 {
	return 1 << 32
}

type HeadersMsg struct {
	HeaderIndices HeaderIndices
	Headers Headers
}

var hdrMsgSSZ = zssz.GetSSZ((*HeadersMsg)(nil))

func (c *HeadersChunk) PushHeaders(logger *log.Logger, recv ReceivePort, headers ...uint32) {
	msg := HeadersMsg{}
	for _, hi := range headers {
		if hi > uint32(len(c.Headers)) {
			logger.Printf("Skipping header %d push, only have %d headers", hi, len(c.Headers))
			continue
		}
		msg.HeaderIndices = append(msg.HeaderIndices, hi)
		msg.Headers = append(msg.Headers, c.Headers[hi])
	}
	var out bytes.Buffer
	if err := zssz.Encode(&out, &msg, hdrMsgSSZ); err != nil {
		logger.Printf("can't encode headers msg: %v", err)
		return
	}
	recv.Send(out.Bytes())
}

type HeadersProducer struct {
	Headers chan *bh.BeaconBlockHeader
	Logger *log.Logger
	Closed bool
}

func simHeader(parent *bh.BeaconBlockHeader) *bh.BeaconBlockHeader {
	if parent == nil {
		// TODO: customize genesis?
		return new(bh.BeaconBlockHeader)
	}
	return &bh.BeaconBlockHeader{
		Slot: parent.Slot + 1,
		ParentRoot: ssz.HashTreeRoot(parent, bh.BeaconBlockHeaderSSZ),
		StateRoot:  core.Root{123},
		BodyRoot:   core.Root{42},
		Signature:  core.BLSSignature{1,2,3,4},
	}
}

func (hp *HeadersProducer) Mock() {
	lookback := 10
	lastX := make([]*bh.BeaconBlockHeader, 0, lookback)
	i := 0
	for {
		if hp.Closed {
			return
		}
		pi := rand.Intn(lookback)
		var parent *bh.BeaconBlockHeader
		if len(lastX) > pi {
			// may be nil
			parent = lastX[len(lastX) - 1 - pi]
		}
		newHeader := simHeader(parent)
		if len(lastX) == lookback {
			copy(lastX, lastX[1:])
			lastX = lastX[:len(lastX) - 1]
		}
		lastX = append(lastX, newHeader)
		i++
		hp.Headers <- newHeader
		time.Sleep(200 * time.Millisecond)
	}
}

func (hp *HeadersProducer) Close() {
	hp.Closed = true
	close(hp.Headers)
}

// The only (synchronous) writer to headers of any chunk
func (hp *HeadersProducer) Process(getChunk common.ChunkGetter, getViewing common.ViewersGetter) {
	for {
		if h, ok := <-hp.Headers; ok {
			chunkID := common.ChunkID(h.Slot / slotsPerChunk)
			ch := getChunk(chunkID)
			chH, ok := ch.(WithHeaders)
			if !ok {
				hp.Logger.Printf("Chunk getter returned unexpected instance: %v", chH)
				continue
			}
			chunk := chH.HeadersChunk()
			index := uint32(len(chunk.Headers))
			// update headers in chunk
			chunk.Headers = append(chunk.Headers, h)
			// update viewers
			for _, v := range getViewing(chunkID) {
				chunk.PushHeaders(hp.Logger, v, index)
			}
		} else {
			break
		}
	}
}
