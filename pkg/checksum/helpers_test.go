package checksum

// helpers for testing
func fullPartialChecksum_01(data []byte, dataLen int) uint32 {
	p := PeChecksum{}
	// calc all the data
	p.partialChecksum_01(data)
	return p.FinalizeChecksum(dataLen)
}

func fullPartialChecksum_02(data []byte, dataLen int) uint32 {
	p := PeChecksum{}
	// calc all the data
	p.partialChecksum_02(data)
	return p.FinalizeChecksum(dataLen)
}

var functions = []struct {
	name string
	f    func(data []byte, dataLen int) uint32
}{
	{"partialChecksum_01", fullPartialChecksum_01},
	{"partialChecksum_02", fullPartialChecksum_02},
}
