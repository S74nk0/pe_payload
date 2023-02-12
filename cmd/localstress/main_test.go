package main

import (
	"fmt"
	"io/ioutil"
	"pe_payload/pkg/appender"
	"runtime"
	"sync"
	"testing"
	"time"
)

func BenchmarkMain(b *testing.B) {

	var wg sync.WaitGroup

	vcbytes, err := ioutil.ReadFile("VC_redist.x64.exe")
	handleErr(err)
	peAppender, err = appender.NewPEDataAppenderDynamic(vcbytes)
	handleErr(err)

	fmt.Println("BEFORE START")
	time.Sleep(time.Second * 5)

	N := 500
	fmt.Println("START stress")

	for i := 0; i < N; i++ {
		wg.Add(1)
		go execAppend(&wg, i)
	}

	fmt.Println("Waiting to finish")
	wg.Wait()
	fmt.Println("DONE stress wait")
	time.Sleep(time.Second * 10)
	fmt.Println("GC call")
	runtime.GC()
	fmt.Println("after GC call")
	time.Sleep(time.Second * 10)
	fmt.Println("DONE")
}
