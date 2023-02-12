package appender

import (
	"encoding/binary"
	"fmt"
	"io"
	"pe_payload/internal/pkg/pe"
	"pe_payload/pkg/checksum"
)

// this is a one time allocated const - only read from this so it is thread safe
var paddZero = []byte("\000")

// TODO test true/false performance difference (I think true will yeild better performance)
var UsePrePadding = true

type PeDataAppender interface {
	Append(w io.Writer, payload []byte) (err error)
	Append0Alloc(w io.Writer, payload, uint32Buffer []byte) (err error)
	FileSize(payloadLen int) int
}

func finalSize(paddingSize, payloadMsgSize, dataLen int) int {
	r := paddingSize + payloadMsgSize + dataLen
	// fmt.Printf("paddingSize %d. Len %d. payloadMsg %d. FINAL %d \n", paddingSize, dataLen, payloadMsgSize, r)
	return r
}

type precalcedChecksum struct {
	checksum.PeChecksum

	// this includes the payload message only without the header
	payloadMsgSize     uint32
	paddingSize        uint32
	newCertTableLength uint32
}

type peDataAppender struct {
	// data splitted in chunks
	data01 []byte
	data02 []byte
	data03 []byte
	data04 []byte

	originalDataLen uint32
	dataLen         uint32

	// PE calculated indexes
	checksumChunkIndex         uint32
	certTableOffsetIndex       uint32
	certTableLengthOffsetIndex uint32

	// data read from the PE headers
	certTableReadOffsetIndex uint32
	// the table lenghts
	certTableReadLen  uint32
	certTableReadLen2 uint32

	// padding
	prePaddingSize uint32

	// payload header
	payloadHeaderSize uint32
}

func (p *peDataAppender) log(tag string) {
	return
	fmt.Println()
	fmt.Println(tag)
	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	fmt.Printf("p.checksumChunkIndex %d\n", p.checksumChunkIndex)
	fmt.Printf("p.certTableOffsetIndex %d\n", p.certTableOffsetIndex)
	fmt.Printf("p.certTableLengthOffsetIndex %d\n", p.certTableLengthOffsetIndex)
	// fmt.Printf("XX %d\n")
	// fmt.Printf("XX %d\n")
	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	fmt.Println()
}

// init functions
func (p *peDataAppender) init01_findPEHeaderAndCalcConstants(data, payloadHeader []byte) (err error) {
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

func (p *peDataAppender) init03_prepareData(origData, payloadHeader []byte, usePrePadding bool) {
	// this is probably not needed anymore since the checksum struct takes care of it
	p.prePaddingSize = 0
	if usePrePadding {
		p.prePaddingSize = uint32(((p.originalDataLen + p.payloadHeaderSize) % 4))
		// fmt.Println("Using pre padding pre pad size = ", p.prePaddingSize)
	}
	p.dataLen = p.originalDataLen + p.payloadHeaderSize + p.prePaddingSize
	// prepare the data, this might have the intermediate pading
	// since we know the size we alocate this ONLY once!!! this is like C calloc so data is zero initialised
	data := make([]byte, p.dataLen, p.dataLen)
	copy(data[:p.originalDataLen], origData)
	copy(data[p.originalDataLen+p.prePaddingSize:], payloadHeader)

	// data splitted in chunks
	p.data01 = data[:p.checksumChunkIndex]
	p.data02 = data[p.checksumChunkIndex+4 : p.certTableLengthOffsetIndex]
	p.data03 = data[p.certTableLengthOffsetIndex+4 : p.certTableReadOffsetIndex]
	p.data04 = data[p.certTableReadOffsetIndex+4:]
}

func (p *peDataAppender) precalcChecksum(payloadMsgSize uint32) precalcedChecksum {
	// from here on we do specific calls depending on the specific appender
	const PAYLOAD_ALIGNMENT = 8
	paddingSize := PAYLOAD_ALIGNMENT - ((p.payloadHeaderSize + payloadMsgSize) % PAYLOAD_ALIGNMENT)
	// Update certification table
	newCertTableLength := p.certTableReadLen + p.payloadHeaderSize + payloadMsgSize + paddingSize + p.prePaddingSize

	checksum := precalcedChecksum{
		payloadMsgSize:     payloadMsgSize,
		paddingSize:        paddingSize,
		newCertTableLength: newCertTableLength,
	}

	newCertTableLengthBuff := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(newCertTableLengthBuff, newCertTableLength)

	// pre-calc checksum
	// from 0 - PE checksum
	// checksum := checksum.PeChecksum{}
	checksum.PartialChecksum(p.data01)
	// skip checksum and continue to calc to first offset
	checksum.PartialChecksum(p.data02)
	// write new checksum table size at FIRST offset
	checksum.PartialChecksum(newCertTableLengthBuff)
	// write rest of data to second certTableLenght
	checksum.PartialChecksum(p.data03)
	// write new checksum table size at SECOND offset
	checksum.PartialChecksum(newCertTableLengthBuff)
	// from PE checksum - Data end
	checksum.PartialChecksum(p.data04)

	return checksum
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

// this is like our own io.WriteTo, it makes only 1 allocation per call with 4bytes
// TODO even though this makes one allocation it can hit the GC hard
func (p *peDataAppender) append(w io.Writer, payload []byte, finalChecksum, newCertTableLength uint32, finalN int) (err error) {
	uint32Buffer := make([]byte, 4, 4)
	var writtenBytes int

	// write until checksum
	n, err := w.Write(p.data01)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	// write checksum
	binary.LittleEndian.PutUint32(uint32Buffer, finalChecksum)
	// err = binary.Write(w, binary.LittleEndian, finalChecksum)
	n, err = w.Write(uint32Buffer)
	if err != nil {
		return
	} else {
		writtenBytes += 4
	}

	// until first table size
	n, err = w.Write(p.data02)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	binary.LittleEndian.PutUint32(uint32Buffer, newCertTableLength)
	// err = binary.Write(w, binary.LittleEndian, newCertTableLength)
	n, err = w.Write(uint32Buffer)
	if err != nil {
		return
	} else {
		writtenBytes += 4
	}

	n, err = w.Write(p.data03)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	// no need to fill the buffer again
	// binary.LittleEndian.PutUint32(uint32Buffer, newCertTableLength)
	// err = binary.Write(w, binary.LittleEndian, newCertTableLength)
	n, err = w.Write(uint32Buffer)
	if err != nil {
		return
	} else {
		writtenBytes += 4
	}

	n, err = w.Write(p.data04)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	// write the payload
	n, err = w.Write(payload)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	paddingBytesSize := finalN - writtenBytes
	// fmt.Printf("paddingBytesSize %d \n", paddingBytesSize)
	for i := 0; i < paddingBytesSize; i++ {
		n, err = w.Write(paddZero)
		if err != nil {
			return
		}
		writtenBytes += n
	}

	if writtenBytes != finalN {
		// TODO this is probably an error
		err = fmt.Errorf("writtenBytes differs from final write size (%d!=%d)", writtenBytes, finalN)
		return
	}

	return
}

func (p *peDataAppender) append_0_alloc(w io.Writer, payload, uint32Buffer []byte, finalChecksum, newCertTableLength uint32, finalN int) (err error) {

	var writtenBytes int

	// write until checksum
	n, err := w.Write(p.data01)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	// write checksum
	binary.LittleEndian.PutUint32(uint32Buffer, finalChecksum)
	// err = binary.Write(w, binary.LittleEndian, finalChecksum)
	n, err = w.Write(uint32Buffer)
	if err != nil {
		return
	} else {
		writtenBytes += 4
	}

	// until first table size
	n, err = w.Write(p.data02)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	binary.LittleEndian.PutUint32(uint32Buffer, newCertTableLength)
	// err = binary.Write(w, binary.LittleEndian, newCertTableLength)
	n, err = w.Write(uint32Buffer)
	if err != nil {
		return
	} else {
		writtenBytes += 4
	}

	n, err = w.Write(p.data03)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	// no need to fill the buffer again
	// binary.LittleEndian.PutUint32(uint32Buffer, newCertTableLength)
	// err = binary.Write(w, binary.LittleEndian, newCertTableLength)
	n, err = w.Write(uint32Buffer)
	if err != nil {
		return
	} else {
		writtenBytes += 4
	}

	n, err = w.Write(p.data04)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	// write the payload
	n, err = w.Write(payload)
	if err != nil {
		return
	} else {
		writtenBytes += n
	}

	paddingBytesSize := finalN - writtenBytes
	// fmt.Printf("paddingBytesSize %d \n", paddingBytesSize)
	for i := 0; i < paddingBytesSize; i++ {
		n, err = w.Write(paddZero)
		if err != nil {
			return
		}
		writtenBytes += n
	}

	if writtenBytes != finalN {
		// TODO this is probably an error
		err = fmt.Errorf("writtenBytes differs from final write size (%d!=%d)", writtenBytes, finalN)
		return
	}

	return
}
