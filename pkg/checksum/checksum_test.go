package checksum

import (
	"crypto/rand"
	"testing"
)

func TestComutativeNaive(t *testing.T) {
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
