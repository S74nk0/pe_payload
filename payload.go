package pe_payload

import (
	"fmt"
)

// payload header starts with and data is from there and after
var APPEND_HEADER = []byte("APPENDNH01\000\000")

func paddedPayload(payload []byte) []byte {
	rem := len(payload) % 4
	if rem == 0 {
		return payload
	}
	padding := 4 - rem
	ret := make([]byte, len(payload)+padding)
	for i := 0; i < len(payload); i++ {
		ret[i] = payload[i]
	}
	return ret
}

func ReadPayload(data []byte) (payload []byte, err error) {
	var headerIndexCheck int
	for i := 0; i < len(data); i++ {
		b := data[i]
		if b == APPEND_HEADER[headerIndexCheck] {
			headerIndexCheck++
			if headerIndexCheck == len(APPEND_HEADER) {
				payload = data[i+1:]
				return
			}
		} else {
			headerIndexCheck = 0
		}
	}

	err = fmt.Errorf("input has no APPENDED Data")
	return
}
