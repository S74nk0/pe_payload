package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"pe_payload/pkg/appender"
	"runtime"
	"sync"
	"time"
)

// io.Writer
type nilWriter struct {
	io.Writer
}

func (w *nilWriter) Write(p []byte) (n int, err error) {
	// fake write
	n = len(p)
	return
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

var payload = []byte("some data n1dfd")

var peAppender appender.PeDataAppender

func execAppend(wg *sync.WaitGroup, i int) {
	good := true
	fmt.Println("start: ", i)
	w := &nilWriter{}
	err := peAppender.Append(w, payload)
	good = err == nil
	fmt.Println("ended: ", i, "good: ", good)
	wg.Done()
}

func main() {

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
