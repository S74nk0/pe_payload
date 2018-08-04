package appender

import (
	"fmt"
	"io"
	"math"
	"pe_payload/pkg/checksum"
	"pe_payload/pkg/payload"
)

type peDataAppenderDynamic struct {
	peDataAppender

	// the smaller the value it will re-init cert table offsets and re-calc init checksum
	// this should be the power of two
	payloadMessageStep uint32
}

func (p *peDataAppenderDynamic) prepare(data, payloadHeader []byte, usePrePadding bool) (err error) {
	err = p.init(data, payloadHeader, usePrePadding)
	if err != nil {
		return
	}
	// we will calc the hash on the fly
	p.log("DYNAMIC")
	return
}

func (p *peDataAppenderDynamic) Append(w io.Writer, payload []byte) (err error) {
	if uint32(len(payload)) > maxDynamicSize {
		err = fmt.Errorf("cannot append paylod with size %d, MAX size is ", len(payload), maxDynamicSize)
		return
	}

	payloadMsgSize := calcPayloadMsgSize(uint32(len(payload)), p.payloadMessageStep)
	fmt.Println("payloadMsgSize: ", payloadMsgSize)
	updateTableAndPreCalcChecksum := p.payloadMsgSize != payloadMsgSize
	if updateTableAndPreCalcChecksum {
		fmt.Println("TABLE AND CHECKSUM UPDATE")
		// from here on we do specific calls depending on the specific appender
		p.calcPayloadMsgPaddings(payloadMsgSize)
		p.updateCertificationTable()

		// pre-calc checksum
		// from 0 - PE checksum
		p.checksum = checksum.PeChecksum{}
		p.checksum.PartialChecksum(p.data[:p.checksumChunkIndex])
		// from PE checksum - Data end
		p.checksum.PartialChecksum(p.data[p.checksumChunkIndex+4:])
	}

	checksum := p.checksum.DeepCopy()
	// calc rest of the checksum
	checksum.PartialChecksum(payload)
	finalChecksum := checksum.FinalizeChecksum(p.finalSize())

	err = p.append(w, payload, finalChecksum)
	return
}

// TODO fix it
func (p *peDataAppenderDynamic) FileSize(len int) int {
	_ = len
	return p.finalSize()
}

func NewPEDataAppenderDynamic(originalData []byte) (ret PeDataAppender, err error) {
	p := peDataAppenderDynamic{}
	p.payloadMessageStep = uint32(math.Pow(2, 8))
	err = p.prepare(originalData, payload.APPEND_HEADER, UsePrePadding)
	if err != nil {
		return
	}

	ret = &p
	return
}
