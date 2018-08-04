package appender

import (
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
		b.Run(fmt.Sprintf("Init_%s", f.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err = f.f(vcbytes)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}

func BenchmarkAppending(b *testing.B) {
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
	// benchmarking
	payload := []byte("sdslkdsfldsfjs s f lfjd ")
	for i := range appenders {
		name := names[i]
		var a PeDataAppender = appenders[i]
		b.Run(fmt.Sprintf("Append_%s", name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := a.Append(nW, payload)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}
