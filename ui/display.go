package ui

import (
	"bytes"
)

type display struct {
	buf      *bytes.Buffer
	renderCh chan []byte
}

func newDisplay() *display {
	return &display{
		buf:      bytes.NewBuffer([]byte{}),
		renderCh: make(chan []byte),
	}
}

func (d *display) Write(p []byte) (n int, err error) {
	d.Clear()
	n, err = d.buf.Write(p)
	d.renderCh <- d.buf.Bytes()
	return
}

func (d *display) Clear() {
	d.buf.Reset()
	d.renderCh <- d.buf.Bytes()
}
