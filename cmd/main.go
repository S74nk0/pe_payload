package main

import (
	"io/ioutil"
	"pe_payload"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	vcbytes, err := ioutil.ReadFile("VC_redist.x64.exe")
	handleErr(err)
	appendData, err := ioutil.ReadFile("append.txt")
	handleErr(err)
	payload, err := pe_payload.NewPayload(appendData)
	handleErr(err)

	_, err = pe_payload.Append(vcbytes, payload)
	handleErr(err)
	pe_payload.Checksum(vcbytes)
}
