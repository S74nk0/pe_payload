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

/*
TODO:
Add random data access test functions.
So make it start from non dword alligned chunks to simulate performance hits
Even though we will call 'PartialChecksum functions' at most 2-3 times (it means 2-3 perf hits)
Implement these functions to see what is the perf hit.
We Will and are using this module with dword alligned memory and this will yield best performance
*/

var functions = []struct {
	name string
	f    func(data []byte, dataLen int) uint32
}{
	{"partialChecksum_01", fullPartialChecksum_01},
	{"partialChecksum_02", fullPartialChecksum_02},
}
