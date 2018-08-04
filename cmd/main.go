package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"pe_payload"
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

	f, err := os.Create("new_golangOut.exe")
	appender, err := pe_payload.NewPEDataAppenderFixed(vcbytes)
	handleErr(err)
	err = appender.Append(f, []byte("some data n1dfd"))
	handleErr(err)

	outnew, err := ioutil.ReadFile("new_golangOut.exe")
	handleErr(err)
	payload, err := pe_payload.ReadPayload(outnew)
	handleErr(err)
	fmt.Println(string(payload))
}
