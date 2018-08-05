package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"pe_payload/pkg/appender"
	"strconv"
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
	a, err := appender.NewPEDataAppenderDynamic(vcbytes)
	handleErr(err)

	fmt.Println("Appender initialized")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		fmt.Println(r.RequestURI)

		fmt.Println(r.Header.Get("Content-Type"))

		// TODO get payload from URL or something
		payload := []byte("some data n1dfd")

		dispositionVal := fmt.Sprintf("attachment; filename=%s", "dl.exe")
		fileSizeVal := strconv.Itoa(a.FileSize(len(payload)))
		//copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
		w.Header().Set("Content-Disposition", dispositionVal)
		// w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Type", "application/vnd.microsoft.portable-executable")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Content-Length", fileSizeVal)

		// application/vnd.microsoft.portable-executable

		// stream most from memory and append new data
		err = a.Append(w, payload)
		if err != nil {
			fmt.Printf("Append err: %s\n", err.Error())
		}
	})
	err = http.ListenAndServe(":8000", nil)

	if err != nil {
		fmt.Println(err)
	}
}
