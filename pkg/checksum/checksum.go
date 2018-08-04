package checksum

import (
	"encoding/binary"
	"math"
)

var top = uint64(math.Pow(2, 32))

// func partialChecksum(checksum uint64, data []byte) (ret uint64, err error) {
// 	// pad data otherwise error
// 	if len(data)%4 != 0 {
// 		err = fmt.Errorf("error partial checksum data has remainder %d (%d=%d mod 4), remainder must be 0", len(data)%4, len(data)%4, len(data))
// 		return
// 	}

// 	// checksum body calc
// 	{
// 		iters := len(data) / 4
// 		for i := 0; i < iters; i++ {
// 			dword := binary.LittleEndian.Uint32(data[i*4:])
// 			checksum = (checksum & 0xffffffff) + uint64(dword) + (checksum >> 32)
// 			if checksum > top {
// 				checksum = (checksum & 0xffffffff) + (checksum >> 32)
// 			}
// 		}
// 	}

// 	// ret prepare
// 	ret = checksum
// 	return
// }

// func finalizeChecksum(initChecksum uint64, fileSize int) (ret uint32) {
// 	var checksum = uint64(initChecksum)

// 	checksum = (checksum & 0xffff) + (checksum >> 16)
// 	checksum = (checksum) + (checksum >> 16)
// 	checksum = checksum & 0xffff
// 	checksum += uint64(fileSize)

// 	// ret prepare
// 	ret = uint32(checksum)
// 	return
// }

func CalcCheckSum(data []byte, PECheckSumIndex uint32) uint32 {
	c := PeChecksum{}
	c.PartialChecksum(data[:PECheckSumIndex])
	// skip PECheckSumIndex dword
	c.PartialChecksum(data[PECheckSumIndex+4:])
	return c.FinalizeChecksum(len(data))
}

const checksumBufferSize = 4

type peChecksumPart struct {
	rem uint8
	// array used because it behaves as value
	b [checksumBufferSize]byte
}

func (p *peChecksumPart) zeroOutRemBytes() {
	p.rem = 0
	p.b[0] = 0
	p.b[1] = 0
	p.b[2] = 0
	p.b[3] = 0
}

func (p *peChecksumPart) fillByte(b byte) bool {
	p.b[p.rem] = b
	p.rem++
	return p.rem%checksumBufferSize == 0
}

func (p *peChecksumPart) dword() uint64 {
	return uint64(uint32(p.b[0]) | uint32(p.b[1])<<8 | uint32(p.b[2])<<16 | uint32(p.b[3])<<24)
}

type PeChecksum struct {
	peChecksumPart
	checksum uint64
}

// linearly calc checksum
func (p *PeChecksum) partialChecksum_01(data []byte) {
	dataLen := len(data)
	written := 0
	// #01 check if we have some bytes from before
	if p.rem != 0 {
		for i := 0; i < dataLen; i++ {
			written++
			if p.fillByte(data[i]) {
				p.calcChecksumBuffer()
				break
			}
		}
	}

	// #02 calc dword body, checksum body calc
	{
		remOff := written
		iters := (dataLen - written) / 4
		written += iters * 4
		for i := 0; i < iters; i++ {
			ind := remOff + i*4
			dword := binary.LittleEndian.Uint32(data[ind:])
			p.checksum = (p.checksum & 0xffffffff) + uint64(dword) + (p.checksum >> 32)
			if p.checksum > top {
				p.checksum = (p.checksum & 0xffffffff) + (p.checksum >> 32)
			}
		}
	}

	// #03 calc rem
	for i := written; i < dataLen; i++ {
		// written++
		if p.fillByte(data[i]) {
			p.calcChecksumBuffer()
		}
	}
}

func (p *PeChecksum) calcChecksumBuffer() {
	dword := p.dword()
	p.checksum = (p.checksum & 0xffffffff) + dword + (p.checksum >> 32)
	if p.checksum > top {
		p.checksum = (p.checksum & 0xffffffff) + (p.checksum >> 32)
	}
	p.zeroOutRemBytes()
}

// linearly calc checksum
func (p *PeChecksum) partialChecksum_02(data []byte) {
	for i := 0; i < len(data); i++ {
		if p.fillByte(data[i]) {
			p.calcChecksumBuffer()
		}
	}
}

// linearly calc checksum. order matters this is not comutative
func (p *PeChecksum) PartialChecksum(data []byte) {
	// TODO test performance probably 02 is slower
	p.partialChecksum_01(data)
	// p.partialChecksum_02(data)
}

func (p *PeChecksum) FinalizeChecksum(fileSize int) (ret uint32) {
	// finalize padded checksum
	if p.rem != 0 {
		p.calcChecksumBuffer()
	}
	p.checksum = (p.checksum & 0xffff) + (p.checksum >> 16)
	p.checksum = (p.checksum) + (p.checksum >> 16)
	p.checksum = p.checksum & 0xffff
	p.checksum += uint64(fileSize)

	// ret prepare
	ret = uint32(p.checksum)
	return
}

func (p *PeChecksum) Reset() {
	p.zeroOutRemBytes()
	p.checksum = 0
}

func (p *PeChecksum) DeepCopy() (ret PeChecksum) {
	// this will copy
	ret = *p
	// // fill data
	// ret.checksum = p.checksum
	// ret.rem = p.rem
	// ret.b[0] = p.b[0]
	// ret.b[1] = p.b[1]
	// ret.b[2] = p.b[2]
	// ret.b[3] = p.b[3]

	return
}
