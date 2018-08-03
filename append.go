package pe_payload

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
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

func CalcCheckSum(data []byte, PECheckSum uint32) uint32 {
	var checksum uint64
	top := uint64(math.Pow(2, 32))
	// checksum body calc
	{
		iters := uint32(len(data) / 4)
		iterSkip := PECheckSum / 4
		for i := uint32(0); i < iters; i++ {
			dword := binary.LittleEndian.Uint32(data[i*4:])
			if i == iterSkip {
				fmt.Printf("CALC CALC %d\n", dword)
				continue
			}
			checksum = (checksum & 0xffffffff) + uint64(dword) + (checksum >> 32)
			if checksum > top {
				checksum = (checksum & 0xffffffff) + (checksum >> 32)
			}
		}
	}
	// remainder check scope
	{
		rem := len(data) % 4
		fmt.Printf("remainder %d\n", rem)
		// last step
		if rem != 0 {
			lastChunk := len(data) / 4
			remBytes := make([]byte, payload_size)
			for i := 0; i < rem; i++ {
				remBytes[i] = data[lastChunk+i]
			}
			dword := binary.LittleEndian.Uint32(remBytes)
			checksum = (checksum & 0xffffffff) + uint64(dword) + (checksum >> 32)
			if checksum > top {
				checksum = (checksum & 0xffffffff) + (checksum >> 32)
			}
		}
	}

	// checksum finalize
	{
		checksum = (checksum & 0xffff) + (checksum >> 16)
		checksum = (checksum) + (checksum >> 16)
		checksum = checksum & 0xffff
		checksum += uint64(len(data))
	}

	fmt.Printf("--- ")
	fmt.Print(checksum)
	fmt.Printf(" ---\n")
	return uint32(checksum)
}

// func

func Checksum(data []byte) (err error) {
	const OPT_CHECKSUM_OFFSET = 88

	peHeaderStart, _, err := Find_PE_Header(data)
	checksum := binary.LittleEndian.Uint32(data[OPT_CHECKSUM_OFFSET+peHeaderStart:])
	fmt.Printf("checksum %d\n", checksum)
	fmt.Println(checksum == 15272560)
	fmt.Printf("checksum END\n")

	// calc checksum
	c := CalcCheckSum(data, peHeaderStart+OPT_CHECKSUM_OFFSET)
	fmt.Println(c)
	fmt.Println(c == 15272560)

	return
}

func Append(data, payload []byte) (out *bytes.Buffer, err error) {
	in := bytes.NewBuffer(data)

	const CERTIFICATE_ENTRY_OFFSET = 148
	const PAYLOAD_ALIGNMENT = 8

	peHeaderStart, peHeaderEnd, err := Find_PE_Header(data)
	_ = peHeaderStart
	// Get PE\0\0 Header signature
	var PE_HEADER = []byte("PE\000\000") // {'P', 'E', '\000', '\000'}
	var peIndexCheck = 0

	var cert_table_length_offset uint32 = CERTIFICATE_ENTRY_OFFSET + 4
	var cert_table_length_offset2 uint32 = CERTIFICATE_ENTRY_OFFSET + peHeaderEnd + 4

PE_HEADER_SEARCH_LOOP:
	for {
		cert_table_length_offset++
		b, err := in.ReadByte()
		if err == io.EOF {
			err2 := fmt.Errorf("Input is not a valid PE Executable")
			return nil, err2
		} else if b == PE_HEADER[peIndexCheck] {
			peIndexCheck++
			if peIndexCheck == len(PE_HEADER) {
				break PE_HEADER_SEARCH_LOOP
			}
		} else {
			peIndexCheck = 0
		}
	}
	if cert_table_length_offset != cert_table_length_offset2 {
		err = fmt.Errorf("wrong offset size %d:%d", cert_table_length_offset, cert_table_length_offset2)
		return
	}
	in.Next(CERTIFICATE_ENTRY_OFFSET)
	cert_table_offset_bytes := in.Next(4)
	cert_table_length_bytes := in.Next(4)

	cert_table_offset := binary.LittleEndian.Uint32(cert_table_offset_bytes)
	cert_table_length := binary.LittleEndian.Uint32(cert_table_length_bytes)

	fmt.Printf("cert_table_offset: %d || cert_table_length: %d \n", cert_table_offset, cert_table_length)

	fmt.Printf("cert_table_offset: %d || cert_table_length_offset: %d \n", cert_table_offset, cert_table_length_offset)

	padding_size := PAYLOAD_ALIGNMENT - (payload_size % PAYLOAD_ALIGNMENT)
	fmt.Printf("padding_size %d \n", padding_size)

	in = bytes.NewBuffer(data)
	in.Next(int(cert_table_offset))

	cert_table_length2 := binary.LittleEndian.Uint32(in.Next(4))
	if cert_table_length != cert_table_length2 {
		return nil, fmt.Errorf("Failed to read certificate table location properly")
	}
	if int(cert_table_offset+cert_table_length) != len(data) {
		return nil, fmt.Errorf("The certificate table is not located at the end of the file!")
	}

	outFile := bytes.NewBuffer(data)
	outFile.Write(payload)
	for i := 0; i < int(padding_size); i++ {
		outFile.WriteByte('\000')
	}

	// Update certification table
	outBytes := outFile.Bytes()
	cert_table_length_new := cert_table_length + payload_size + padding_size
	// certTableNewBytes := make([]byte, 4)
	// binary.LittleEndian.PutUint32(certTableNewBytes, cert_table_length_new)
	// for i := uint32(0); i < 4; i++ {
	// 	outBytes[cert_table_length_offset+i] = certTableNewBytes[i]
	// }
	// for i := uint32(0); i < 4; i++ {
	// 	outBytes[cert_table_offset+i] = certTableNewBytes[i]
	// }
	binary.LittleEndian.PutUint32(outBytes[cert_table_length_offset:], cert_table_length_new)
	binary.LittleEndian.PutUint32(outBytes[cert_table_offset:], cert_table_length_new)

	// out->pubseekpos(cert_table_length_offset);
	// out->sputn(reinterpret_cast<char*>(&cert_table_length), sizeof(DWORD));
	// out->pubseekpos(cert_table_offset);
	// out->sputn(reinterpret_cast<char*>(&cert_table_length), sizeof(DWORD));

	err = ioutil.WriteFile("golangOut.exe", outBytes, 0777)
	if err != nil {
		return
	}

	p, _ := ReadPayload(outBytes)
	fmt.Println(p)

	return
}
