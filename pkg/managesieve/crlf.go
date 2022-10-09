package managesieve

import (
	"bytes"
)

var (
	crlf = []byte{'\r', '\n'}
	cr   = []byte{'\r'}
	lf   = []byte{'\n'}
)

func EnsureCRLF(data []byte) []byte {
	data = bytes.ReplaceAll(data, lf, crlf)
	data = bytes.ReplaceAll(data, []byte{cr[0], cr[0]}, cr)
	data = bytes.ReplaceAll(data, cr, crlf)
	data = bytes.ReplaceAll(data, []byte{lf[0], lf[0]}, lf)

	return data
}
