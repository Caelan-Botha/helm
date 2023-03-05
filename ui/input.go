package ui

import "bytes"

type input struct {
	buf      *bytes.Buffer
	renderCh chan []byte
}

func newInput() *input {
	return &input{
		buf:      bytes.NewBuffer([]byte{}),
		renderCh: make(chan []byte),
	}
}

func (i *input) Write(p []byte) (n int, err error) {
	n, err = i.buf.Write(p)
	i.renderCh <- i.buf.Bytes()
	return
}
func (i *input) Read(p []byte) (n int, err error) {
	n, err = i.buf.Read(p)
	i.renderCh <- i.buf.Bytes()
	return
}

func (i *input) Backspace() {
	if i.buf.Len() == 0 {
		return
	}
	i.buf.Truncate(i.buf.Len() - 1)
	i.renderCh <- i.buf.Bytes()
}

func (i *input) Enter(sendTo *bytes.Buffer) (n int64, err error) {
	i.buf.Write([]byte{'\n'})
	return sendTo.ReadFrom(i)
}
