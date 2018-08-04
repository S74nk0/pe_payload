package checksum

import (
	"encoding/binary"
	"fmt"
	"math"
)

var top = uint64(math.Pow(2, 32))

func partialChecksum(checksum uint64, data []byte) (ret uint64, err error) {
	// pad data otherwise error
	if len(data)%4 != 0 {
		err = fmt.Errorf("error partial checksum data has remainder %d (%d=%d mod 4), remainder must be 0", len(data)%4, len(data)%4, len(data))
		return
	}

	// checksum body calc
	{
		iters := len(data) / 4
		for i := 0; i < iters; i++ {
			dword := binary.LittleEndian.Uint32(data[i*4:])
			checksum = (checksum & 0xffffffff) + uint64(dword) + (checksum >> 32)
			if checksum > top {
				checksum = (checksum & 0xffffffff) + (checksum >> 32)
			}
		}
	}

	// ret prepare
	ret = checksum
	return
}

func finalizeChecksum(initChecksum uint64, fileSize int) (ret uint32) {
	var checksum = uint64(initChecksum)

	checksum = (checksum & 0xffff) + (checksum >> 16)
	checksum = (checksum) + (checksum >> 16)
	checksum = checksum & 0xffff
	checksum += uint64(fileSize)

	// ret prepare
	ret = uint32(checksum)
	return
}

func finalChecksum(partialChecksumS uint64, payload []byte, fileSize int) (ret uint32, err error) {
	fmt.Printf("fileSize %d. fileSize rem %d\n", fileSize, fileSize%4)
	checksum, err := partialChecksum(partialChecksumS, payload)
	if err != nil {
		return
	}
	// checksum finalize
	ret = finalizeChecksum(checksum, fileSize)
	return
}

// func CalcCheckSum(data []byte, PECheckSum uint32) uint32 {
// 	var checksum uint64
// 	// checksum body calc
// 	{
// 		iters := uint32(len(data) / 4)
// 		iterSkip := PECheckSum / 4
// 		for i := uint32(0); i < iters; i++ {
// 			dword := binary.LittleEndian.Uint32(data[i*4:])
// 			if i == iterSkip {
// 				fmt.Printf("CALC CALC %d\n", dword)
// 				continue
// 			}
// 			checksum = (checksum & 0xffffffff) + uint64(dword) + (checksum >> 32)
// 			if checksum > top {
// 				checksum = (checksum & 0xffffffff) + (checksum >> 32)
// 			}
// 		}
// 	}
// 	// remainder check scope
// 	{
// 		rem := len(data) % 4
// 		fmt.Printf("remainder %d\n", rem)
// 		// last step
// 		if rem != 0 {
// 			lastChunk := len(data) / 4
// 			remBytes := make([]byte, 4)
// 			for i := 0; i < rem; i++ {
// 				remBytes[i] = data[lastChunk+i]
// 			}
// 			dword := binary.LittleEndian.Uint32(remBytes)
// 			checksum = (checksum & 0xffffffff) + uint64(dword) + (checksum >> 32)
// 			if checksum > top {
// 				checksum = (checksum & 0xffffffff) + (checksum >> 32)
// 			}
// 		}
// 	}

// 	// checksum finalize
// 	ret := finalizeChecksum(uint32(checksum), len(data))
// 	return ret
// }

type PeChecksum struct {
	checksum       uint64
	rem            int
	checksumBuffer []byte
}

func (p *PeChecksum) zeroOutRemBytes() {
	p.rem = 0
	// TODO unroll?
	p.checksumBuffer[0] = 0
	p.checksumBuffer[1] = 0
	p.checksumBuffer[2] = 0
	p.checksumBuffer[3] = 0

	// // for now keep the roll
	// for i := 0; i < 4; i++ {
	// 	p.remBytes[i] = 0
	// }
}

// // MAKE SURE TO NEVER PASS data bigger than the remBytesBuff
// func (p *PeChecksum) fillRem(data []byte) {
// 	// for now keep the roll
// 	for i := 0; i < len(data); i++ {
// 		p.remBytes[p.rem+i] = data[i]
// 	}
// }

func (p *PeChecksum) fillByte(b byte) bool {
	p.checksumBuffer[p.rem] = b
	p.rem++
	return p.rem%4 == 0
}

// // linearly calc checksum
// func (p *PeChecksum) PartialChecksum(data []byte) {
// 	dataLen := len(data)
// 	dataRem := dataLen % 4
// 	remOff := p.rem % 4
// 	written := 0
// 	// #01 check if we have some bytes from before
// 	if remOff != 0 {
// 		p.fillRem(data[:remOff])
// 		p.checksum, _ = partialChecksum(p.checksum, p.remBytes)
// 		p.zeroOutRemBytes()
// 		remOff++
// 		written = remOff + 1
// 	}

// 	if written >= dataLen {
// 		return
// 	}

// 	// #02 calc dword body
// 	{
// 		start := remOff
// 		remOff = (dataLen - remOff) % 4
// 		end := dataLen - remOff
// 		p.checksum, _ = partialChecksum(p.checksum, data[start:end])
// 	}

// 	// #03 calc rem
// 	p.rem = dataLen - written
// 	{
// 		start := remOff
// 		remOff = (dataLen - remOff) % 4
// 		end := dataLen - remOff
// 		p.checksum, _ = partialChecksum(p.checksum, data[start:end])
// 	}
// }

func (p *PeChecksum) calcChecksumBuffer() {
	dword := binary.LittleEndian.Uint32(p.checksumBuffer)
	p.checksum = (p.checksum & 0xffffffff) + uint64(dword) + (p.checksum >> 32)
	if p.checksum > top {
		p.checksum = (p.checksum & 0xffffffff) + (p.checksum >> 32)
	}
	p.zeroOutRemBytes()
}

// linearly calc checksum
func (p *PeChecksum) PartialChecksum(data []byte) {
	for i := 0; i < len(data); i++ {
		if p.fillByte(data[i]) {
			//p.checksum, _ = partialChecksum(p.checksum, p.remBytes)
			p.calcChecksumBuffer()
		}
	}
}

func (p *PeChecksum) FinalizeChecksum(fileSize int) (ret uint32) {
	// TODO finalize padded checksum
	if p.rem != 0 {
		p.calcChecksumBuffer()
	}
	return finalizeChecksum(p.checksum, fileSize)
}

func (p *PeChecksum) DeepCopy(fileSize int) (ret PeChecksum) {
	ret = NewPeChecksum()
	// fill data
	ret.checksum = p.checksum
	ret.rem = p.rem
	ret.checksumBuffer[0] = p.checksumBuffer[0]
	ret.checksumBuffer[1] = p.checksumBuffer[1]
	ret.checksumBuffer[2] = p.checksumBuffer[2]
	ret.checksumBuffer[3] = p.checksumBuffer[3]
	return
}

func NewPeChecksum() PeChecksum {
	return PeChecksum{
		checksumBuffer: make([]byte, 4, 4),
	}
}
