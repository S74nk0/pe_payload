package appender

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"testing"
)

func BenchmarkInitializersDefault(b *testing.B) {
	vcbytes, err := ioutil.ReadFile("./testingFolder/VC_redist.x64.exe")
	if err != nil {
		b.Error(err)
	}
	for _, f := range newFunctions {
		b.Run(fmt.Sprintf("%s", f.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err = f.f(vcbytes)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}

func BenchmarkAppendingFixedLoad(b *testing.B) {
	vcbytes, err := ioutil.ReadFile("./testingFolder/VC_redist.x64.exe")
	if err != nil {
		b.Error(err)
	}
	// init appenders
	appenders := make([]PeDataAppender, len(newFunctions), len(newFunctions))
	names := make([]string, len(newFunctions), len(newFunctions))
	for i, f := range newFunctions {
		names[i] = f.name
		appenders[i], err = f.f(vcbytes)
		if err != nil {
			b.Error(err)
		}
	}
	// prepare payload
	payload := make([]byte, 512)
	rand.Read(payload)
	// benchmarking
	for i := 0; i < len(appenders); i++ {
		name := names[i]
		var a PeDataAppender = appenders[i]
		b.Run(fmt.Sprintf("%s_%d", name, len(payload)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := a.Append(nW, payload)
				if err != nil {
					b.Error(err)
				}
				// a.Append(nW, payload)
			}
		})
	}
	// // reverse
	// for i := len(appenders) - 1; i >= 0; i-- {
	// 	name := names[i]
	// 	var a PeDataAppender = appenders[i]
	// 	b.Run(fmt.Sprintf("%s_R_%d", name, len(payload)), func(b *testing.B) {
	// 		for i := 0; i < b.N; i++ {
	// 			err := a.Append(nW, payload)
	// 			if err != nil {
	// 				b.Error(err)
	// 			}
	// 		}
	// 	})
	// }

}

func BenchmarkAppendingDifferentLoad(b *testing.B) {
	vcbytes, err := ioutil.ReadFile("./testingFolder/VC_redist.x64.exe")
	if err != nil {
		b.Error(err)
	}
	// init appenders
	appenders := make([]PeDataAppender, len(newFunctions), len(newFunctions))
	names := make([]string, len(newFunctions), len(newFunctions))
	for i, f := range newFunctions {
		names[i] = f.name
		appenders[i], err = f.f(vcbytes)
		if err != nil {
			b.Error(err)
		}
	}
	// prepare payload
	payloadFull := make([]byte, 512)
	rand.Read(payloadFull)
	// benchmarking
	for lenOff := 0; lenOff <= 512; lenOff += 51 {
		payload := payloadFull[:lenOff]
		for i := 0; i < len(appenders); i++ {
			name := names[i]
			var a PeDataAppender = appenders[i]
			b.Run(fmt.Sprintf("%s_%d", name, len(payload)), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					err := a.Append(nW, payload)
					if err != nil {
						b.Error(err)
					}
					// a.Append(nW, payload)
				}
			})
		}
		// // reverse
		// for i := len(appenders) - 1; i >= 0; i-- {
		// 	name := names[i]
		// 	var a PeDataAppender = appenders[i]
		// 	b.Run(fmt.Sprintf("%s_R_%d", name, len(payload)), func(b *testing.B) {
		// 		for i := 0; i < b.N; i++ {
		// 			err := a.Append(nW, payload)
		// 			if err != nil {
		// 				b.Error(err)
		// 			}
		// 		}
		// 	})
		// }
	}
}
