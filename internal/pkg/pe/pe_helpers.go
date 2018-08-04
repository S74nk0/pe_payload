package pe

import "fmt"

// PE consts
const (
	OPT_CHECKSUM_OFFSET      = 88
	CERTIFICATE_ENTRY_OFFSET = 148
)

func Find_PE_Header(data []byte) (peHeaderStart, peHeaderEnd uint32, err error) {
	// Get PE\0\0 Header signature
	var PE_HEADER = []byte("PE\000\000") // {'P', 'E', '\000', '\000'}
	var peIndexCheck = 0
	for i := 0; i < len(data); i++ {
		peHeaderEnd++
		b := data[i]
		if b == PE_HEADER[peIndexCheck] {
			peIndexCheck++
			if peIndexCheck == len(PE_HEADER) {
				peHeaderStart = peHeaderEnd - uint32(len(PE_HEADER))
				return
			}
		} else {
			peIndexCheck = 0
		}
	}
	err = fmt.Errorf("input data is not a valid PE Executable")
	return
}

func ChecksumChunkIndex(peHeaderStart uint32) uint32 { return peHeaderStart + OPT_CHECKSUM_OFFSET }
func CertTableOffsetIndex(peHeaderEnd uint32) uint32 { return peHeaderEnd + CERTIFICATE_ENTRY_OFFSET }
func CertTableLengthOffsetIndex(peHeaderEnd uint32) uint32 {
	return peHeaderEnd + CERTIFICATE_ENTRY_OFFSET + 4
}
