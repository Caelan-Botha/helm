package helm

import "bytes"

// ? textBuffer
// ? =========================================================================================

type TextBuffer struct {
	renderCh chan []byte
	buf      *bytes.Buffer
}

func NewTextBuffer(out chan []byte) *TextBuffer {
	return &TextBuffer{
		renderCh: out,
		buf:      bytes.NewBuffer([]byte{}),
	}
}

func (t *TextBuffer) String() string {
	return t.buf.String()
}
func (t *TextBuffer) Backspace() {
	t.buf.Truncate(t.buf.Len() - 1)
	t.renderCh <- t.buf.Bytes()
}
func (t *TextBuffer) Enter() {
	t.renderCh <- t.buf.Bytes()
}

// write
func (t *TextBuffer) write(p []byte) (n int, err error) {
	n, err = t.buf.Write(p)
	t.renderCh <- t.buf.Bytes()
	return
}
func (t *TextBuffer) writeString(s string) (n int, err error) {
	n, err = t.buf.WriteString(s)
	t.renderCh <- t.buf.Bytes()
	return
}
func (t *TextBuffer) writeRune(r rune) (n int, err error) {
	n, err = t.buf.WriteRune(r)
	t.renderCh <- t.buf.Bytes()
	return
}

// read
func (t *TextBuffer) read(p []byte) (n int, err error) {
	return t.buf.Read(p)
}
func (t *TextBuffer) readString(delim byte) (line string, err error) {
	return t.buf.ReadString(delim)
}
func (t *TextBuffer) readRune() (r rune, size int, err error) {
	return t.buf.ReadRune()
}
