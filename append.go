package pe_payload

import "io"

type PeDataAppender interface {
	Append(w io.Writer, payload []byte) (err error)
	FileSize() int
}
