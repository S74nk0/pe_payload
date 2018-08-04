package checksum

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func BenchmarkPartialChecksumFunctions(b *testing.B) {
	const maxSize = 60 * 1000000
	// step by 10
	for dataLen := (100 * 4); dataLen < maxSize; dataLen *= 10 {
		data := make([]byte, dataLen)
		rand.Read(data)
		for _, f := range functions {
			b.Run(fmt.Sprintf("bench_%s_fsize_%d", f.name, dataLen), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					f.f(data, dataLen)
				}
			})
		}
	}
}
