package checksum

import (
	"crypto/rand"
	"testing"
)

// this is not really a test
// func TestComutativeNaive(t *testing.T) {
// 	dataLen := 100 * 4
// 	data := make([]byte, dataLen)
// 	rand.Read(data)

// 	p := PeChecksum{}
// 	p.PartialChecksum(data)
// 	refChecksum := p.FinalizeChecksum(dataLen)

// 	for i := 1; i < 100; i++ {
// 		// p.Reset()
// 		p = PeChecksum{}

// 		p.PartialChecksum(data[i:])
// 		p.PartialChecksum(data[:i])
// 		checksum := p.FinalizeChecksum(dataLen)
// 		if checksum != refChecksum {
// 			t.Error("checksum is different. We assume not comutative")
// 		}
// 	}
// }

func TestPartialChecksumFunctions(t *testing.T) {
	dataLen := 100 * 4
	data := make([]byte, dataLen)
	rand.Read(data)

	p := PeChecksum{}
	p.PartialChecksum(data)
	refChecksum := p.FinalizeChecksum(dataLen)

	for i := 1; i < 100; i++ {
		// p.Reset()
		p = PeChecksum{}

		p.PartialChecksum(data[i:])
		p.PartialChecksum(data[:i])
		checksum := p.FinalizeChecksum(dataLen)
		if checksum != refChecksum {
			t.Error("checksum is different. We assume not comutative")
		}
	}
}

func TestPartialChecksumFunctionsEquals(t *testing.T) {
	const maxSize = 60 * 1000000
	// step by 10
	for dataLen := (100 * 4); dataLen < maxSize; dataLen *= 10 {
		data := make([]byte, dataLen)
		rand.Read(data)

		checksums := make([]uint32, len(functions), len(functions))
		for i, f := range functions {
			checksums[i] = f.f(data, dataLen)
		}
		// compare checksums
		{
			for i := 0; i < len(checksums); i++ {
				for j := i + 1; j < len(checksums); j++ {
					if checksums[i] != checksums[j] {
						t.Error("checksums are not equal")
					}
				}
			}
		}
	}
}

// TODO add actual PE file checksum tests
