package appender

import (
	"encoding/binary"
	"fmt"
	"io"
	"pe_payload/internal/pkg/pe"
	"pe_payload/pkg/checksum"
)

// TODO test true/false performance difference (I think true will yeild better performance)
var UsePrePadding = true

type PeDataAppender interface {
	Append(w io.Writer, payload []byte) (err error)
	FileSize(payloadLen int) int
}

type peDataAppender struct {
	// data is the prepared bytes part
	data            []byte
	originalDataLen uint32

	// PE calculated indexes
	checksumChunkIndex         uint32
	certTableOffsetIndex       uint32
	certTableLengthOffsetIndex uint32

	// data read from the PE headers
	certTableReadOffsetIndex uint32
	// the table lenghts
	certTableReadLen  uint32
	certTableReadLen2 uint32

	// padding sizes
	paddingSize    uint32
	prePaddingSize uint32

	// this includes the payload message only without the header
	payloadMsgSize    uint32
	payloadHeaderSize uint32

	// base checksum
	checksum checksum.PeChecksum
}

func (p *peDataAppender) log(tag string) {
	fmt.Println()
	fmt.Println(tag)
	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	fmt.Printf("p.checksumChunkIndex %d\n", p.checksumChunkIndex)
	fmt.Printf("p.certTableOffsetIndex %d\n", p.certTableOffsetIndex)
	fmt.Printf("p.certTableLengthOffsetIndex %d\n", p.certTableLengthOffsetIndex)
	// fmt.Printf("p.partialChecksum %d\n", p.partialChecksum)
	fmt.Printf("p.paddingSize %d\n", p.paddingSize)
	fmt.Printf("p.payloadMsgSize %d\n", p.payloadMsgSize)
	// fmt.Printf("XX %d\n")
	// fmt.Printf("XX %d\n")
	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	fmt.Println()
}

func (p *peDataAppender) finalSize() int {
	fmt.Printf("paddingSize %d. Len %d. payloadMsg %d \n", p.paddingSize, uint32(len(p.data)), p.payloadMsgSize)
	r := p.paddingSize + uint32(len(p.data)) + p.payloadMsgSize
	return int(r)
}

// init functions
func (p *peDataAppender) init01_findPEHeaderAndCalcConstants(data []byte, payloadHeader []byte) (err error) {
	peHeaderStart, peHeaderEnd, err := pe.Find_PE_Header(data)
	if err != nil {
		return
	}

	// set known calculated constants
	p.checksumChunkIndex = pe.ChecksumChunkIndex(peHeaderStart)
	p.certTableOffsetIndex = pe.CertTableOffsetIndex(peHeaderEnd)
	p.certTableLengthOffsetIndex = pe.CertTableLengthOffsetIndex(peHeaderEnd)
	p.payloadHeaderSize = uint32(len(payloadHeader))
	p.originalDataLen = uint32(len(data))

	return
}

func (p *peDataAppender) init02_assertAppendPossibleAndReadTableValues(data []byte) (err error) {
	// check and make sure we can do the appending
	p.certTableReadOffsetIndex = binary.LittleEndian.Uint32(data[p.certTableOffsetIndex:])
	p.certTableReadLen = binary.LittleEndian.Uint32(data[p.certTableLengthOffsetIndex:])
	p.certTableReadLen2 = binary.LittleEndian.Uint32(data[p.certTableReadOffsetIndex:])

	if p.certTableReadLen != p.certTableReadLen2 {
		err = fmt.Errorf("failed to read certificate table location properly")
		return
	}
	if (p.certTableReadOffsetIndex + p.certTableReadLen) != p.originalDataLen {
		err = fmt.Errorf("the certificate table is not located at the end of the file!")
		return
	}
	return
}

func (p *peDataAppender) init03_prepareData(data, payloadHeader []byte, usePrePadding bool) {
	// this is probably not needed anymore since the checksum struct takes care of it
	p.prePaddingSize = 0
	if usePrePadding {
		p.prePaddingSize = uint32(((p.originalDataLen + p.payloadHeaderSize) % 4))
		fmt.Println("Using pre padding pre pad size = ", p.prePaddingSize)
	}
	dataLen := p.originalDataLen + p.payloadHeaderSize + p.prePaddingSize
	// prepare the data, this might have the intermediate pading
	// since we know the size we alocate this ONLY once!!! this is like C calloc so data is zero initialised
	p.data = make([]byte, dataLen, dataLen)
	copy(p.data[:p.originalDataLen], data)
	copy(p.data[p.originalDataLen+p.prePaddingSize:], payloadHeader)
}

func (p *peDataAppender) calcPayloadMsgPaddings(payloadMsgSize uint32) {
	const PAYLOAD_ALIGNMENT = 8
	p.payloadMsgSize = payloadMsgSize
	p.paddingSize = PAYLOAD_ALIGNMENT - ((p.payloadHeaderSize + payloadMsgSize) % PAYLOAD_ALIGNMENT)
}

// IMPORTANT when calling this function make sure you have initialized and when payload size changes call calcPayloadMsgPaddings
func (p *peDataAppender) updateCertificationTable() {
	// Update certification table
	newCertTableLength := p.certTableReadLen + p.payloadHeaderSize + p.payloadMsgSize + p.paddingSize + p.prePaddingSize
	binary.LittleEndian.PutUint32(p.data[p.certTableLengthOffsetIndex:], newCertTableLength)
	binary.LittleEndian.PutUint32(p.data[p.certTableReadOffsetIndex:], newCertTableLength)
}

func (p *peDataAppender) init(data, payloadHeader []byte, usePrePadding bool) (err error) {
	err = p.init01_findPEHeaderAndCalcConstants(data, payloadHeader)
	if err != nil {
		return
	}
	err = p.init02_assertAppendPossibleAndReadTableValues(data)
	if err != nil {
		return
	}
	p.init03_prepareData(data, payloadHeader, usePrePadding)
	return
}

func (p *peDataAppender) append(w io.Writer, payload []byte, finalChecksum uint32) (err error) {
	var finalN int
	n, err := w.Write(p.data[:p.checksumChunkIndex])
	if err != nil {
		return
	}
	finalN += n
	err = binary.Write(w, binary.LittleEndian, finalChecksum)
	if err != nil {
		return
	}
	finalN += 4
	n, err = w.Write(p.data[p.checksumChunkIndex+4:])
	if err != nil {
		return
	}
	finalN += n
	n, err = w.Write(payload)
	if err != nil {
		return
	}
	finalN += n

	paddingBytesSize := p.paddingSize + p.payloadMsgSize - uint32(len(payload))
	paddingBytes := make([]byte, paddingBytesSize)
	n, err = w.Write(paddingBytes)
	if err != nil {
		return
	}
	finalN += n

	if finalN != p.finalSize() {
		// TODO this is probably an error
		fmt.Println("finalN differs from final write size (%d!=%d)", finalN, p.finalSize())
	}

	return
}
