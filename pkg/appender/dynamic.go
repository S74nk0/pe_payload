package appender

import (
	"fmt"
	"io"
	"math"
	"pe_payload/pkg/payload"
)

// TODO maybe make this a variable
// 0.5 MB should be more than enough
const maxDynamicSize = 1000000 / 2

func calcPayloadMsgSize(payloadMsgSize, payloadMessageStep uint32) uint32 {
	mult := (payloadMsgSize / payloadMessageStep)
	if (payloadMsgSize % payloadMessageStep) != 0 {
		mult++
	}
	ret := mult * payloadMessageStep

	// fmt.Printf("## (%d / %d) + %d = %d ##\n", payloadMsgSize, payloadMessageStep, payloadMessageStep, ret)
	return ret
}

// with pre-calculated sizes

type peDataAppenderDynamic struct {
	peDataAppender

	// the smaller the value it will re-init cert table offsets and re-calc init checksum
	// this should be the power of two
	payloadMessageStep uint32
	buckets            map[uint32]*precalcedChecksum

	preInitFirst uint32
}

// this function gets the precalculated checksum or precalculates one on the fly if it is missing and returns it
func (p *peDataAppenderDynamic) getChecksumLazy(keyPayloadMsgSize uint32) precalcedChecksum {
	precalcedChecksum, exists := p.buckets[keyPayloadMsgSize]
	if !exists || precalcedChecksum == nil {
		payloadMsgSize := keyPayloadMsgSize
		// pre-calc checksum
		c := p.precalcChecksum(payloadMsgSize)
		p.buckets[payloadMsgSize] = &c
		return c

	}

	// this will create a copy, we want a copy
	return *precalcedChecksum
}

func (p *peDataAppenderDynamic) prepare(data, payloadHeader []byte, usePrePadding bool) (err error) {
	err = p.init(data, payloadHeader, usePrePadding)
	if err != nil {
		return
	}

	// pre calc buckets
	for i := uint32(0); i < p.preInitFirst; i++ {
		payloadMsgSize := p.payloadMessageStep * i
		p.getChecksumLazy(payloadMsgSize)
	}

	p.log("DYNAMIC_BUCKETS")
	return
}

func (p *peDataAppenderDynamic) Append(w io.Writer, payload []byte) (err error) {
	if uint32(len(payload)) > maxDynamicSize {
		err = fmt.Errorf("cannot append paylod with size %d, MAX size is %d", len(payload), maxDynamicSize)
		return
	}
	payloadMsgSize := calcPayloadMsgSize(uint32(len(payload)), p.payloadMessageStep)
	checksum := p.getChecksumLazy(payloadMsgSize)
	// calc rest of the checksum
	checksum.PartialChecksum(payload)
	finalN := finalSize(int(checksum.paddingSize), int(checksum.payloadMsgSize), int(p.dataLen))
	finalChecksum := checksum.FinalizeChecksum(finalN)

	err = p.append(w, payload, finalChecksum, checksum.newCertTableLength, finalN) // fastest #1
	return
}

func (p *peDataAppenderDynamic) Append0Alloc(w io.Writer, payload, uint32Buffer []byte) (err error) {
	if uint32(len(payload)) > maxDynamicSize {
		err = fmt.Errorf("cannot append paylod with size %d, MAX size is %d", len(payload), maxDynamicSize)
		return
	}
	payloadMsgSize := calcPayloadMsgSize(uint32(len(payload)), p.payloadMessageStep)
	checksum := p.getChecksumLazy(payloadMsgSize)
	// calc rest of the checksum
	checksum.PartialChecksum(payload)
	finalN := finalSize(int(checksum.paddingSize), int(checksum.payloadMsgSize), int(p.dataLen))
	finalChecksum := checksum.FinalizeChecksum(finalN)

	err = p.append_0_alloc(w, payload, uint32Buffer, finalChecksum, checksum.newCertTableLength, finalN)
	return
}

// TODO fix it
func (p *peDataAppenderDynamic) FileSize(l int) int {
	payloadMsgSize := calcPayloadMsgSize(uint32(l), p.payloadMessageStep)
	checksum := p.getChecksumLazy(payloadMsgSize)
	finalN := finalSize(int(checksum.paddingSize), int(checksum.payloadMsgSize), int(p.dataLen))
	return finalN
}

func NewPEDataAppenderDynamic(originalData []byte) (ret PeDataAppender, err error) {
	p := peDataAppenderDynamic{
		buckets:      make(map[uint32]*precalcedChecksum),
		preInitFirst: 9,
	}
	p.payloadMessageStep = uint32(math.Pow(2, 6))
	err = p.prepare(originalData, payload.APPEND_HEADER, UsePrePadding)
	if err != nil {
		return
	}

	ret = &p
	return
}
