package ui

import (
	"bytes"
)

type Formatter interface {
	Format(p []byte) ([]byte, error)
}

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

func (d *display) WriteFormatted(p []byte, formatter Formatter) (n int, err error) {
	d.Clear()

	pCpy := make([]byte, len(p))
	copy(pCpy, p)

	var formattedBytes []byte
	formattedBytes, err = formatter.Format(pCpy)
	if err != nil {
		return 0, err
	}
	return d.Write(formattedBytes)
}

func (d *display) Clear() {
	d.buf.Reset()
	d.renderCh <- d.buf.Bytes()
}
