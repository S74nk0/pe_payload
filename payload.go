package pe_payload

import (
	"bytes"
	"fmt"
	"io"
)

var APPEND_HEADER = []byte("APPEND_01\000\000")

// make this payload dynamic in the future

const defaultPayloadSize = 512

var payload_size = uint32(len(APPEND_HEADER) + defaultPayloadSize)

type Payload []byte

func NewPayload(data []byte) (p []byte, err error) {
	if len(data) > defaultPayloadSize {
		err = fmt.Errorf("unable to inject payload of size %d bytes. Supported MAX size is %d bytes", len(data), defaultPayloadSize)
		return
	}

	// zero valued byte slice
	p = make([]byte, payload_size)

	// buf := bytes.NewBuffer(p)
	// err = binary.Write(buf, binary.LittleEndian, APPEND_HEADER)
	// if err != nil {
	// 	err = fmt.Errorf("failed to write APPEND_HEADER (size %d, %s). binary.Write failed: %s", len(APPEND_HEADER), string(APPEND_HEADER), err.Error())
	// 	return
	// }
	// err = binary.Write(buf, binary.LittleEndian, data)
	// if err != nil {
	// 	err = fmt.Errorf("failed to write data (size %d, %s). binary.Write failed: %s", len(data), string(data), err.Error())
	// 	return
	// }
	for i := 0; i < len(APPEND_HEADER); i++ {
		p[i] = APPEND_HEADER[i]
	}
	for i := 0; i < len(data); i++ {
		p[len(APPEND_HEADER)+i] = data[i]
	}

	return
}

func ReadPayload(data []byte) (payload []byte, err error) {
	r := bytes.NewBuffer(data)
	var headerIndexCheck int
APPEND_HEADER_SEARCH_LOOP:
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			err2 := fmt.Errorf("Input has no APPENDED Data")
			return nil, err2
		} else if b == APPEND_HEADER[headerIndexCheck] {
			headerIndexCheck++
			if headerIndexCheck == len(APPEND_HEADER) {
				break APPEND_HEADER_SEARCH_LOOP
			}
		} else {
			headerIndexCheck = 0
		}
	}
	payload = r.Next(defaultPayloadSize)
	return
}
