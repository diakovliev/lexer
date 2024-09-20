package xio

import "bytes"

type buffer = bytes.Buffer

func newBuffer(buf []byte) *buffer {
	return bytes.NewBuffer(buf)
}
