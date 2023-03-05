package ui

import (
	"bytes"
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
)

// ? ui
// ? =========================================================================================

//go:embed title.txt
var title string

type UI struct {
	// internal
	app            fyne.App
	window         fyne.Window
	title          *widget.TextGrid
	displaySection *widget.TextGrid
	inputSection   *widget.TextGrid
	display        *display
	input          *input

	// external
	externalBuffer *bytes.Buffer
	Reader         Reader
	Writer         Writer
}

func NewUI() *UI {
	externalBuffer := bytes.NewBuffer([]byte{})
	display := newDisplay()
	input := newInput()
	return &UI{
		app:            app.New(),
		display:        display,
		input:          input,
		externalBuffer: externalBuffer,
		Reader:         Reader{externalBuffer: externalBuffer},
		Writer:         Writer{display: display},
	}
}

func (u *UI) Start() {
	// create the window
	u.window = u.app.NewWindow("helm")

	u.title = widget.NewTextGrid()
	u.title.SetText(title)

	u.displaySection = widget.NewTextGrid()
	u.inputSection = widget.NewTextGrid()

	// keypress handler
	onTypedKey := func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEnter || e.Name == fyne.KeyReturn {
			// handle enter pressed
			_, err := u.input.Enter(u.externalBuffer)
			if err != nil {
				log.Fatal("failed to handle enter key press: ", err)
			}
		}
		if e.Name == fyne.KeyBackspace {
			u.input.Backspace()
		}
	}

	// rune display handler
	onTypedRune := func(r rune) {
		_, err := u.input.Write([]byte(string(r)))
		if err != nil {
			log.Fatal(err)
		}
	}

	u.window.Canvas().SetOnTypedKey(onTypedKey)
	u.window.Canvas().SetOnTypedRune(onTypedRune)

	// render loop
	go u.inputRenderLoop()
	go u.displayRenderLoop()

	// create container with wrapped layout
	u.window.SetContent(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(1000, 100)), u.title, u.displaySection, u.inputSection),
	)

	u.window.ShowAndRun()
	defer u.window.Close()
}

// ? render
// ? =========================================================================================

func (u *UI) RenderInput(b []byte) {
	u.inputSection.SetText(string(b))
}

func (u *UI) inputRenderLoop() {
	for {
		b := <-u.input.renderCh
		u.RenderInput(b)
	}
}

func (u *UI) RenderDisplay(b []byte) {
	u.displaySection.SetText(string(b))
}

func (u *UI) displayRenderLoop() {
	for {
		b := <-u.display.renderCh
		u.RenderDisplay(b)

	}
}

// ? reader
// ? =========================================================================================

type Reader struct {
	externalBuffer *bytes.Buffer
}

func (r Reader) Read(p []byte) (n int, err error) {
	return r.externalBuffer.Read(p)
}

// ? writer
// ? =========================================================================================

type Writer struct {
	display *display
}

func (w Writer) Write(p []byte) (n int, err error) {
	return w.display.Write(p)
}
