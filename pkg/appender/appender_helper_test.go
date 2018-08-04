package appender

import "io"

// io.Writer
type nilWriter struct {
	io.Writer
}

func (w *nilWriter) Write(p []byte) (n int, err error) {
	// fake write
	n = len(p)
	return
}

var nW io.Writer = &nilWriter{}

var newFunctions = []struct {
	name string
	f    func(originalData []byte) (ret PeDataAppender, err error)
}{
	{"NewPEDataAppenderFixed", NewPEDataAppenderFixed},
	{"NewPEDataAppenderDynamic", NewPEDataAppenderDynamic},
	{"NewPEDataAppenderDynamicBuckets", NewPEDataAppenderDynamicBuckets},
}
