package appender

import (
	"fmt"
	"io"
	"math"
	"pe_payload/pkg/checksum"
	"pe_payload/pkg/payload"
)

// with pre-calculated sizes

type peDataAppenderDynamicBuckets struct {
	peDataAppender

	// the smaller the value it will re-init cert table offsets and re-calc init checksum
	// this should be the power of two
	payloadMessageStep uint32
	buckets            map[uint32]*checksum.PeChecksum

	preInitFirst uint32
}

// this function gets the precalculated checksum or precalculates one on the fly if it is missing
// also it sets the table correctly
func (p *peDataAppenderDynamicBuckets) getChecksumLazy(keyPayloadMsgSize uint32) checksum.PeChecksum {
	precalcedChecksum, exists := p.buckets[keyPayloadMsgSize]
	if !exists || precalcedChecksum == nil {
		payloadMsgSize := keyPayloadMsgSize

		p.calcPayloadMsgPaddings(payloadMsgSize)
		p.updateCertificationTable()

		// pre-calc checksum
		// from 0 - PE checksum
		c := checksum.PeChecksum{}
		c.PartialChecksum(p.data[:p.checksumChunkIndex])
		// from PE checksum - Data end
		c.PartialChecksum(p.data[p.checksumChunkIndex+4:])

		p.buckets[payloadMsgSize] = &c
		return c.DeepCopy()

	}

	// fmt.Println("payloadMsgSize: ", keyPayloadMsgSize)
	updateCertTable := p.payloadMsgSize != keyPayloadMsgSize
	if updateCertTable {
		// fmt.Println("TABLE CERT UPDATE")
		// from here on we do specific calls depending on the specific appender
		p.calcPayloadMsgPaddings(keyPayloadMsgSize)
		p.updateCertificationTable()
	}

	return precalcedChecksum.DeepCopy()
}

func (p *peDataAppenderDynamicBuckets) prepare(data, payloadHeader []byte, usePrePadding bool) (err error) {
	err = p.init(data, payloadHeader, usePrePadding)
	if err != nil {
		return
	}

	// pre calc buckets
	for i := uint32(0); i < p.preInitFirst; i++ {
		payloadMsgSize := p.payloadMessageStep * i

		// from here on we do specific calls depending on the specific appender
		p.calcPayloadMsgPaddings(payloadMsgSize)
		p.updateCertificationTable()

		// pre-calc checksum
		// from 0 - PE checksum
		c := checksum.PeChecksum{}
		c.PartialChecksum(p.data[:p.checksumChunkIndex])
		// from PE checksum - Data end
		c.PartialChecksum(p.data[p.checksumChunkIndex+4:])

		p.buckets[payloadMsgSize] = &c
	}

	p.log("DYNAMIC_BUCKETS")
	return
}

func (p *peDataAppenderDynamicBuckets) Append(w io.Writer, payload []byte) (err error) {
	if uint32(len(payload)) > maxDynamicSize {
		err = fmt.Errorf("cannot append paylod with size %d, MAX size is %d", len(payload), maxDynamicSize)
		return
	}
	payloadMsgSize := calcPayloadMsgSize(uint32(len(payload)), p.payloadMessageStep)
	checksum := p.getChecksumLazy(payloadMsgSize)
	// calc rest of the checksum
	checksum.PartialChecksum(payload)
	finalChecksum := checksum.FinalizeChecksum(p.finalSize())

	err = p.append(w, payload, finalChecksum)
	return
}

// TODO fix it
func (p *peDataAppenderDynamicBuckets) FileSize(len int) int {
	_ = len
	return p.finalSize()
}

func NewPEDataAppenderDynamicBuckets(originalData []byte) (ret PeDataAppender, err error) {
	p := peDataAppenderDynamicBuckets{
		buckets:      make(map[uint32]*checksum.PeChecksum),
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
