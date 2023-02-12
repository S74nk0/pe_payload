package main

import (
	"fmt"
	"io"
	"net"
	"sync"
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
		// panic(err)
	}
}

func downloadFromUrlToNil(address string) (string, bool) {
	output := &nilWriter{}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Sprint("Error while downloading", address, "-", err), false
	}

	defer conn.Close()

	_, err = fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	if err != nil {
		return fmt.Sprint("Error while downloading", address, "-", err), false
	}

	n, err := io.Copy(output, conn)
	if err != nil {
		return fmt.Sprint("Error while downloading", address, "-", err), false
	}

	return fmt.Sprint(n, "bytes downloaded."), true
}

func execWget(wg *sync.WaitGroup, i int) {
	ret := "__"
	ok := true
	defer func() {
		if !ok {
			fmt.Println("ended: ", i, "ret: ", ret, "ok ", ok)
		}

		wg.Done()
	}()
	// fmt.Println("start: ", i)
	ret, ok = downloadFromUrlToNil("127.0.0.1:8000")
}

func main() {

	var wg sync.WaitGroup

	N := 1500
	fmt.Println("START stress")

	for i := 0; i < N; i++ {
		wg.Add(1)
		go execWget(&wg, i)
	}

	fmt.Println("Waiting to finish")
	wg.Wait()
	fmt.Println("DONE stress")
}
