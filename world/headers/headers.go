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
	"sort"
	"time"
)

const slotsPerChunk = 16

type WithHeaders interface {
	HeadersChunk() *HeadersChunk
}

type HeadersChunk struct {
	Headers []*HeaderData
}

type HeadersRequest struct {
	HighestKnown uint32
	Wanted HeaderIndices
}

var hdrRequestSSZ = zssz.GetSSZ((*HeadersRequest)(nil))

func (c *HeadersChunk) HandleRequest(logger *log.Logger, recv ReceivePort, msg []byte) {
	var req HeadersRequest
	if err := zssz.Decode(bytes.NewReader(msg), uint64(len(msg)), &req, hdrRequestSSZ); err != nil {
		logger.Printf("malformed request: %v", err)
		return
	}
	if !sort.IsSorted(req.Wanted) {
		logger.Println("request is malformed; wanted indices are not sorted, ignoring it.")
		return
	}
	max := uint32(len(c.Headers))
	if max < req.HighestKnown {
		logger.Println("request is malformed; claiming higher known than available, ignoring it.")
		return
	}
	s := uint32(len(req.Wanted))
	e := uint32(len(req.Wanted)) + (max - req.HighestKnown)
	res := make([]uint32, e, e)
	copy(res[:], req.Wanted)
	j := s
	for i := req.HighestKnown; i < max; i++ {
		res[j] = i
		j++
	}
	c.PushHeaders(logger, recv, res...)
}

type HeaderIndices []uint32

func (indices HeaderIndices) Len() int {
	return len(indices)
}

func (indices HeaderIndices) Less(i, j int) bool {
	return indices[i] < indices[j]
}

func (indices HeaderIndices) Swap(i, j int) {
	indices[i], indices[j] = indices[j], indices[i]
}

func (*HeaderIndices) Limit() uint64 {
	return 1 << 10
}

var hdrIndicesSSZ = zssz.GetSSZ((*HeaderIndices)(nil))

type HeaderData struct {
	Header *bh.BeaconBlockHeader
	Root core.Root
}

type Headers []*HeaderData

func (*Headers) Limit() uint64 {
	return 1 << 10
}

type HeadersRes struct {
	Indices HeaderIndices
	Headers Headers
}

var hdrResSSZ = zssz.GetSSZ((*HeadersRes)(nil))

func (c *HeadersChunk) PushHeaders(logger *log.Logger, recv ReceivePort, headers ...uint32) {
	msg := HeadersRes{}
	for _, hi := range headers {
		if hi > uint32(len(c.Headers)) {
			logger.Printf("Skipping header %d push, only have %d headers", hi, len(c.Headers))
			continue
		}
		msg.Indices = append(msg.Indices, hi)
		msg.Headers = append(msg.Headers, c.Headers[hi])
	}
	var out bytes.Buffer
	if err := zssz.Encode(&out, &msg, hdrResSSZ); err != nil {
		logger.Printf("can't encode headers msg: %v", err)
		return
	}
	recv.Send(out.Bytes())
}

type HeadersProducer struct {
	Headers chan *HeaderData
	Logger *log.Logger
	Closed bool
}

func simHeader(parent *HeaderData) *HeaderData {
	if parent == nil {
		// TODO: customize genesis?
		h := new(bh.BeaconBlockHeader)
		return &HeaderData{
			Header: h,
			Root:   ssz.HashTreeRoot(h, bh.BeaconBlockHeaderSSZ),
		}
	}
	h := &bh.BeaconBlockHeader{
		Slot: parent.Header.Slot + 1,
		ParentRoot: parent.Root,
		StateRoot:  core.Root{123},
		BodyRoot:   core.Root{42},
		Signature:  core.BLSSignature{1,2,3,4},
	}
	return &HeaderData{
		Header: h,
		Root:   ssz.HashTreeRoot(h, bh.BeaconBlockHeaderSSZ),
	}
}

func (hp *HeadersProducer) Mock() {
	lookback := 10
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
			chunkID := common.ChunkID(h.Header.Slot / slotsPerChunk)
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
