package tests

import (
	"bytes"
	"io"
)

type capturingWriter struct {
	io.Writer
	out    io.Writer
	buffer *bytes.Buffer
}

func newCapturingWriter(out io.Writer) *capturingWriter {
	return &capturingWriter{Writer: out, out: out, buffer: new(bytes.Buffer)}
}

func (c *capturingWriter) startCapturing() {
	c.Writer = c.buffer
}
