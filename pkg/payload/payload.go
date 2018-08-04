package payload

import (
	"fmt"
)

// payload header starts with and data is from there and after
var OLD_SERVER_APPEND_HEADER = []byte("NHPAY:")

var DEFAULT_NO_ENCODING = []byte("APPENDNH01\000\000")

// TODO temp default one
var APPEND_HEADER = OLD_SERVER_APPEND_HEADER

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
