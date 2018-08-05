package appender

import (
	"fmt"
	"io"
	"pe_payload/pkg/payload"
)

type peDataAppenderFixed struct {
	peDataAppender

	// base checksum
	checksum precalcedChecksum
}

func (p *peDataAppenderFixed) prepare(data, payloadHeader []byte, payloadMsgSize uint32, usePrePadding bool) (err error) {
	err = p.init(data, payloadHeader, usePrePadding)
	if err != nil {
		return
	}
	// pre-calc checksum
	p.checksum = p.precalcChecksum(payloadMsgSize)
	p.log("FIXED")
	return
}

func (p *peDataAppenderFixed) Append(w io.Writer, payload []byte) (err error) {
	if uint32(len(payload)) > p.checksum.payloadMsgSize {
		err = fmt.Errorf("cannot append paylod with size %d, MAX size is %d", len(payload), p.checksum.payloadMsgSize)
		return
	}

	// deep copy
	checksum := p.checksum
	// calc rest of the checksum
	checksum.PartialChecksum(payload)
	finalN := finalSize(int(checksum.paddingSize), int(checksum.payloadMsgSize), int(p.dataLen))
	finalChecksum := checksum.FinalizeChecksum(finalN)

	// fmt.Printf("Append equals %t", checksum == p.checksum)

	err = p.append(w, payload, finalChecksum, checksum.newCertTableLength, finalN)
	return
}

func (p *peDataAppenderFixed) FileSize(l int) int {
	_ = l
	finalN := finalSize(int(p.checksum.paddingSize), int(p.checksum.payloadMsgSize), int(p.dataLen))
	return finalN
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
