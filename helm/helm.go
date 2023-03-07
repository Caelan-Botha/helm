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
type CommandFunc func(helm *Helm) error
type SubCommands map[string]Route

// Route represents a Command and its related sub commands
type Route struct {
	mainCommand CommandFunc
	subCommands map[string]Route
}

func (r Route) SubCommands() map[string]Route {
	return r.subCommands
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

//func (h *Helm) AddRoute(name string, mainCommandFunc CommandFunc, subCommandFuncsMap SubCommands) error {
//	if _, exists := h.routes[name]; exists {
//		return fmt.Errorf("a Route with name: %s already exists", name)
//	}
//
//	h.routes[name] = newRoute(mainCommandFunc, subCommandFuncsMap)
//
//	return nil
//}

func (h *Helm) RegisterRoute(name string, route Route) error {
	if _, exists := h.routes[name]; exists {
		return fmt.Errorf("a Route with name: %s already exists", name)
	}

	h.routes[name] = route

	return nil
}

func (h *Helm) CurrentCommand() Command {
	if h.currentCmd.name != "" {
		return h.currentCmd
	}
	return ZeroCommand()
}

func (h *Helm) Start() {
	go h.inputLoop()

	<-h.quitCh
}

func (h *Helm) Quit() {
	close(h.quitCh)
	var empty Helm
	h.quitCh = empty.quitCh
	h.term = empty.term
	h.in = empty.in
	h.out = empty.out
	h.routes = empty.routes
	h.currentCmd = empty.currentCmd
}

// ? input
// ? ========================================================================================================================================================

func (h *Helm) inputLoop() {
	for {
		time.Sleep(1 * time.Millisecond)
		reader := bufio.NewReader(h.in)
		line, err := reader.ReadString('\n')
		if err != nil && errors.Is(err, io.EOF) {
			time.Sleep(99 * time.Millisecond)
			continue
		}
		if err != nil {
			log.Fatal("failed to read from input loop", err)
		}
		h.currentCmd, err = newCommand([]byte(line), h.routes)
		if err != nil {
			h.OutputError(err)
			continue
		}
		//h.routeCommand()
		h.routeCommand()
	}
}

func (h *Helm) routeCommand() {
	command := h.currentCmd
	route, exists := h.routes[command.name]
	if !exists {
		// Command isn't registered
		h.OutputError(fmt.Errorf("unknown command: %s", command.name))
		return
	}
	h.recurs(command, route)
}

func (h *Helm) recurs(command Command, route Route) {
	fmt.Println("recursing: ", command.name, command.subCommand, route)
	if command.subCommand == nil {
		err := route.mainCommand(h)
		if err != nil {
			h.OutputError(err)
			return
		}
		return
	}
	if _, exists := route.subCommands[command.subCommand.name]; !exists {
		h.OutputError(errors.New("Cannot find route for sub-command: " + command.subCommand.name))
		return
	}
	h.recurs(*command.subCommand, route.subCommands[command.subCommand.name])
}

// ? output
// ? ========================================================================================================================================================

func (h *Helm) OutputString(o string) {
	var sb strings.Builder
	sb.WriteString(o)
	sb.WriteRune('\n')
	_, err := h.out.Write([]byte(sb.String()))
	if err != nil {
		log.Fatalf("failed to write to out: %v", err)
	}
}

// ? error
// ? ========================================================================================================================================================

func (h *Helm) OutputError(e error) {
	var sb strings.Builder
	sb.WriteString("ERROR: ")
	sb.WriteString(e.Error())
	_, err := h.out.Write([]byte(sb.String()))
	if err != nil {
		log.Fatalf("failed to write to out: %v", err)
	}
}
