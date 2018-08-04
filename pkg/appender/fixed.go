package appender

import (
	"fmt"
	"io"
	"pe_payload/pkg/checksum"
	"pe_payload/pkg/payload"
)

type peDataAppenderFixed struct {
	peDataAppender
}

func (p *peDataAppenderFixed) prepare(data, payloadHeader []byte, payloadMsgSize uint32, usePrePadding bool) (err error) {
	err = p.init(data, payloadHeader, usePrePadding)
	if err != nil {
		return
	}
	// from here on we do specific calls depending on the specific appender
	p.calcPayloadMsgPaddings(payloadMsgSize)
	p.updateCertificationTable()

	// pre-calc checksum
	// from 0 - PE checksum
	p.checksum = checksum.PeChecksum{}
	p.checksum.PartialChecksum(p.data[:p.checksumChunkIndex])
	// from PE checksum - Data end
	p.checksum.PartialChecksum(p.data[p.checksumChunkIndex+4:])

	p.log("FIXED")
	return
}

func (p *peDataAppenderFixed) Append(w io.Writer, payload []byte) (err error) {
	if uint32(len(payload)) > p.payloadMsgSize {
		err = fmt.Errorf("cannot append paylod with size %d, MAX size is ", len(payload), p.payloadMsgSize)
		return
	}

	checksum := p.checksum.DeepCopy()
	// calc rest of the checksum
	checksum.PartialChecksum(payload)
	finalChecksum := checksum.FinalizeChecksum(p.finalSize())

	err = p.append(w, payload, finalChecksum)
	return
}

func (p *peDataAppenderFixed) FileSize(len int) int {
	_ = len
	return p.finalSize()
}

func NewPEDataAppenderFixed(originalData []byte) (ret PeDataAppender, err error) {
	defaultPayloadSize := uint32(512)
	p := peDataAppenderFixed{}
	err = p.prepare(originalData, payload.APPEND_HEADER, defaultPayloadSize, false)
	if err != nil {
		return
	}

	ret = &p
	return
}
