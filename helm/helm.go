package helm

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/Caelan-Botha/helm/ui"
	"io"
	"log"
	"strings"
	"time"
)

// CommandFunc
// types for grouping commands and their sub-commands
// the main Command will be a Command executor that will branch to the sub-commands if needed.
// this is done by registering the pair of commands/sub-commands as a Route on the schell
// the main Command calls sub-commands if needed
type CommandFunc func(helm *Helm, subCommands SubCommandFuncsMap) error
type SubCommandFuncsMap map[string]CommandFunc

// Route represents a Command and its related sub commands
// mainCommandFunc: is the function that is run when the Command is called, this should contain the branching logic for sub commands eg. a switch statement
type Route struct {
	mainCommandFunc    CommandFunc
	subCommandFuncsMap map[string]CommandFunc
}

func (r Route) SubCommands() map[string]CommandFunc {
	return r.subCommandFuncsMap
}

func newRoute(mainCommandFunc CommandFunc, subCommandFuncsMap SubCommandFuncsMap) Route {
	return Route{
		mainCommandFunc:    mainCommandFunc,
		subCommandFuncsMap: subCommandFuncsMap,
	}
}

// ? Helm
// ? ========================================================================================================================================================

type Helm struct {
	in   io.Reader
	out  io.Writer
	term *ui.UI

	routes map[string]Route

	currentCmd Command

	quitCh chan struct{}
}

func NewHelmUI() (*Helm, *ui.UI) {
	term := ui.NewUI()
	return &Helm{
		in:     term.Reader,
		out:    term.Writer,
		term:   term,
		routes: make(map[string]Route),
		quitCh: make(chan struct{}),
	}, term
}

func NewHelm(in io.Reader, out io.Writer) *Helm {
	return &Helm{
		in:     in,
		out:    out,
		routes: make(map[string]Route),
		quitCh: make(chan struct{}),
	}
}

func (t *Helm) AddRoute(name string, mainCommandFunc CommandFunc, subCommandFuncsMap SubCommandFuncsMap) error {
	if _, exists := t.routes[name]; exists {
		return fmt.Errorf("a Route with name: %s already exists", name)
	}

	t.routes[name] = newRoute(mainCommandFunc, subCommandFuncsMap)

	return nil
}

func (t *Helm) CurrentCommand() Command {
	if t.currentCmd.name != "" {
		return t.currentCmd
	}
	return ZeroCommand()
}

func (t *Helm) Start() {
	go t.inputLoop()

	<-t.quitCh
}

func (t *Helm) Quit() {
	close(t.quitCh)
	var empty Helm
	t.quitCh = empty.quitCh
	t.term = empty.term
	t.in = empty.in
	t.out = empty.out
	t.routes = empty.routes
	t.currentCmd = empty.currentCmd
}

// ? routing
// ? ========================================================================================================================================================

func (t *Helm) GetRoute(name string) (Route, error) {
	if _, exists := t.routes[name]; !exists {
		return Route{}, errors.New("route doesnt exist")
	}
	return t.routes[name], nil
}

// Route the Command to the appropriate handler
func (t *Helm) routeCommand() {
	if _, exists := t.routes[t.currentCmd.name]; !exists {
		// Command isn't registered
		t.OutputError(fmt.Errorf("unknow command: %s", t.currentCmd.name))
		return
	}
	// execute Command main func which will handle any sub commands
	err := t.routes[t.currentCmd.name].mainCommandFunc(t, t.routes[t.currentCmd.name].subCommandFuncsMap)
	if err != nil {
		t.OutputError(err)
	}
}

// ? input
// ? ========================================================================================================================================================

func (t *Helm) inputLoop() {
	for {
		time.Sleep(1 * time.Millisecond)
		reader := bufio.NewReader(t.in)
		line, err := reader.ReadString('\n')
		if err != nil && errors.Is(err, io.EOF) {
			time.Sleep(99 * time.Millisecond)
			continue
		}
		if err != nil {
			log.Fatal("failed to read from input loop", err)
		}
		t.currentCmd, err = newCommand([]byte(line))
		t.routeCommand()
	}
}

// ? output
// ? ========================================================================================================================================================

func (t *Helm) OutputString(o string) {
	var sb strings.Builder
	sb.WriteString(o)
	sb.WriteRune('\n')
	_, err := t.out.Write([]byte(sb.String()))
	if err != nil {
		log.Fatalf("failed to write to out: %v", err)
	}
}

// ? error
// ? ========================================================================================================================================================

func (t *Helm) OutputError(e error) {
	var sb strings.Builder
	sb.WriteString("ERROR: ")
	sb.WriteString(e.Error())
	_, err := t.out.Write([]byte(sb.String()))
	if err != nil {
		log.Fatalf("failed to write to out: %v", err)
	}
}
