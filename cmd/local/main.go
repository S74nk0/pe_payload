package main

import (
	"io/ioutil"
	"os"
	"pe_payload/pkg/appender"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	vcbytes, err := ioutil.ReadFile("VC_redist.x64.exe")
	// vcbytes, err := ioutil.ReadFile("VSCodeSetup-ia32-1.25.1.exe")
	handleErr(err)
	// appendData, err := ioutil.ReadFile("append.txt")
	// handleErr(err)
	// payload, err := pe_payload.NewPayload(appendData)
	// handleErr(err)

	payloadData := []byte("some data n1dfd")
	// payloadData := make([]byte, 512)
	// rand.Read(payloadData)

	f, err := os.Create("new_golangOutStatic.exe")
	peAppender, err := appender.NewPEDataAppenderFixed(vcbytes)
	handleErr(err)
	err = peAppender.Append(f, payloadData)
	handleErr(err)
	f.Close()

	f, err = os.Create("new_golangOutDynamic.exe")
	peAppender, err = appender.NewPEDataAppenderDynamicBuckets(vcbytes)
	handleErr(err)
	err = peAppender.Append(f, payloadData)
	handleErr(err)
	f.Close()

	// outnew, err := ioutil.ReadFile("new_golangOut.exe")
	// handleErr(err)
	// payload, err := payload.ReadPayload(outnew)
	// handleErr(err)
	// fmt.Println(string(payload))
}
