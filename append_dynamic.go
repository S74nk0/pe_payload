package pe_payload

// import (
// 	"bytes"
// 	"encoding/binary"
// 	"fmt"
// 	"io"
// )

// type peDataAppenderDynamic struct {
// 	// data is the prepared bytes part
// 	data                       []byte
// 	checksumChunkIndex         uint32
// 	certTableOffsetIndex       uint32
// 	certTableLengthOffsetIndex uint32
// 	partialChecksum            uint32
// 	paddingSize                uint32

// 	// this includes the payload message only without the header
// 	payloadMsgSize uint32
// }

// func (p *peDataAppenderDynamic) log() {
// 	fmt.Println()
// 	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
// 	fmt.Printf("p.checksumChunkIndex %d\n", p.checksumChunkIndex)
// 	fmt.Printf("p.certTableOffsetIndex %d\n", p.certTableOffsetIndex)
// 	fmt.Printf("p.certTableLengthOffsetIndex %d\n", p.certTableLengthOffsetIndex)
// 	fmt.Printf("p.partialChecksum %d\n", p.partialChecksum)
// 	fmt.Printf("p.paddingSize %d\n", p.paddingSize)
// 	fmt.Printf("p.payloadMsgSize %d\n", p.payloadMsgSize)
// 	// fmt.Printf("XX %d\n")
// 	// fmt.Printf("XX %d\n")
// 	fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
// 	fmt.Println()
// }

// func (p *peDataAppenderDynamic) finalSize() int {
// 	fmt.Printf("paddingSize %d. Len %d. payloadMsg %d \n", p.paddingSize, uint32(len(p.data)), p.payloadMsgSize)
// 	r := p.paddingSize + uint32(len(p.data)) + p.payloadMsgSize
// 	return int(r)
// }

// func (p *peDataAppenderDynamic) prepare(data []byte, payloadHeader []byte, payloadMsgSize uint32) (err error) {
// 	// PE consts
// 	const OPT_CHECKSUM_OFFSET = 88
// 	const CERTIFICATE_ENTRY_OFFSET = 148
// 	const PAYLOAD_ALIGNMENT = 8

// 	peHeaderStart, peHeaderEnd, err := find_PE_Header(data)
// 	if err != nil {
// 		return
// 	}

// 	p.checksumChunkIndex = peHeaderStart + OPT_CHECKSUM_OFFSET
// 	p.certTableOffsetIndex = peHeaderEnd + CERTIFICATE_ENTRY_OFFSET
// 	p.certTableLengthOffsetIndex = peHeaderEnd + CERTIFICATE_ENTRY_OFFSET + 4
// 	p.payloadMsgSize = payloadMsgSize

// 	prepData := bytes.NewBuffer(data)
// 	prePadding := uint32(4 - ((len(data) + len(payloadHeader)) % 4))
// 	fmt.Printf("prePadding %d \n", prePadding)
// 	for i := uint32(0); i < prePadding; i++ {
// 		fmt.Printf("prePadding %d:%d \n", i, prePadding)
// 		prepData.WriteByte('\000')
// 	}
// 	prepData.Write(payloadHeader)

// 	p.paddingSize = PAYLOAD_ALIGNMENT - ((uint32(len(payloadHeader)) + payloadMsgSize) % PAYLOAD_ALIGNMENT)

// 	cert_table_offset_index := peHeaderEnd + CERTIFICATE_ENTRY_OFFSET + 0
// 	cert_table_length_offset_index := peHeaderEnd + CERTIFICATE_ENTRY_OFFSET + 4

// 	cert_table_offset := binary.LittleEndian.Uint32(data[cert_table_offset_index:])
// 	cert_table_length := binary.LittleEndian.Uint32(data[cert_table_length_offset_index:])

// 	// Update certification table
// 	p.data = prepData.Bytes()
// 	cert_table_length_new := cert_table_length + uint32(len(payloadHeader)) + payloadMsgSize + p.paddingSize + prePadding
// 	binary.LittleEndian.PutUint32(p.data[cert_table_length_offset_index:], cert_table_length_new)
// 	binary.LittleEndian.PutUint32(p.data[cert_table_offset:], cert_table_length_new)

// 	// from 0 - PE checksum
// 	checksum, err := partialChecksum(0, p.data[:p.checksumChunkIndex])
// 	if err != nil {
// 		return
// 	}
// 	// from PE checksum - Data end
// 	checksum, err = partialChecksum(checksum, p.data[p.checksumChunkIndex+4:])
// 	if err != nil {
// 		return
// 	}
// 	p.partialChecksum = checksum

// 	p.log()
// 	return
// }

// func (p *peDataAppenderDynamic) Append(w io.Writer, payload []byte) (err error) {
// 	// pad the payload to calc the sum
// 	paddedPayload := paddedPayload(payload)
// 	var finalN int
// 	finalChecksum, err := finalChecksum(p.partialChecksum, paddedPayload, p.finalSize())
// 	if err != nil {
// 		return
// 	}

// 	n, err := w.Write(p.data[:p.checksumChunkIndex])
// 	if err != nil {
// 		return
// 	}
// 	finalN += n
// 	err = binary.Write(w, binary.LittleEndian, finalChecksum)
// 	if err != nil {
// 		return
// 	}
// 	finalN += 4
// 	n, err = w.Write(p.data[p.checksumChunkIndex+4:])
// 	if err != nil {
// 		return
// 	}
// 	finalN += n
// 	n, err = w.Write(payload)
// 	if err != nil {
// 		return
// 	}
// 	paddingBytesSize := p.paddingSize + p.payloadMsgSize - uint32(len(payload))
// 	paddingBytes := make([]byte, paddingBytesSize)
// 	n, err = w.Write(paddingBytes)
// 	if err != nil {
// 		return
// 	}

// 	return
// }

// func NewpeDataAppenderDynamicDynamic(originalData []byte) (ret PeDataAppender, err error) {
// 	defaultPayloadSize := uint32(512)
// 	p := peDataAppenderDynamic{}
// 	err = p.prepare(originalData, APPEND_HEADER, defaultPayloadSize)
// 	if err != nil {
// 		return
// 	}

// 	ret = &p
// 	return
// }
