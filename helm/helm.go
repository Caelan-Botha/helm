package helm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

// ? Command
// ? ========================================================================================================================================================

func ZeroCommand() Command {
	var cmd Command
	return cmd
}

// Command represents the structure of a commands name with args and flags, as well as a pointer to the sub commands
type Command struct {
	name       string
	args       map[string][]string // args are split by delimiter ',' eg. arg=123,555,etc.
	flags      map[byte]struct{}   // this is an efficient way of implementing a set, only use the keys (struct{} allocates 0 memory)
	subCommand *Command
}

func newEmptyCommand() *Command {
	return &Command{
		args:  make(map[string][]string, 0),
		flags: make(map[byte]struct{}, 0),
	}
}

func (c Command) Name() string {
	return c.name
}

func (c Command) Arg(argName string) []string {
	if arg, exists := c.args[argName]; exists {
		return arg
	}
	return nil
}

func (c Command) Flag(flag byte) bool {
	if _, exists := c.flags[flag]; exists {
		return true
	}
	return false
}

func (c Command) SubCommand() Command {
	if c.subCommand != nil {
		return *c.subCommand
	}
	return ZeroCommand()
}

func (c Command) HasSubCommand() bool {
	if c.subCommand != nil {
		return true
	}
	return false
}

// CommandFunc
// types for grouping commands and their sub-commands
// the main Command will be a Command executor that will branch to the sub-commands if needed.
// this is done by registering the pair of commands/sub-commands as a Route on the schell
// the main Command calls sub-commands via the schell dependency injection
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
	in  io.Reader
	out io.Writer

	routes map[string]Route

	currentCmd *Command

	quitCh chan struct{}
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

func (t *Helm) GetRoute(name string) (Route, error) {
	if _, exists := t.routes[name]; !exists {
		return Route{}, errors.New("route doesnt exist")
	}
	return t.routes[name], nil
}

func (t *Helm) GetCurrentCommand() Command {
	if t.currentCmd != nil {
		return *t.currentCmd
	}
	return ZeroCommand()
}

func (t *Helm) Start() {
	go t.inputLoop()

	<-t.quitCh
}

// ? input
// ? ========================================================================================================================================================

func (t *Helm) inputLoop() {
	parseDoneCh := make(chan struct{})
	for {
		time.Sleep(1 * time.Millisecond)
		reader := bufio.NewReader(t.in)
		line, err := reader.ReadString('\n')
		if err != nil && errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			log.Fatal("failed to read from input loop", err)
		}
		go t.parseCommand(line, parseDoneCh)
		<-parseDoneCh
		go t.routeCommand()
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
	sb.WriteRune('\n')
	_, err := t.out.Write([]byte(sb.String()))
	if err != nil {
		log.Fatalf("failed to write to out: %v", err)
	}
}

// ? parsing
// ? ========================================================================================================================================================
// handle input from Command line
func (t *Helm) parseCommand(text string, parseDoneCh chan struct{}) {
	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)
	commandSlc := strings.Split(strings.Trim(text, " "), " ")
	// reset current Command
	t.currentCmd = newEmptyCommand()
	// recursively build Command
	recurs(commandSlc, t.currentCmd, false)
	parseDoneCh <- struct{}{}
}

// recursively builds the Command struct passed in as a pointer
func recurs(commandSlc []string, curr *Command, prev bool) {
	if len(commandSlc) == 0 {
		return
	}
	v := commandSlc[0]
	bytes := []byte(v)
	if len(bytes) == 0 {
		return
	}
	switch bytes[0] {
	case '-':
		// handle "-" flags eg -v
		for _, flag := range bytes[1:] {
			curr.flags[flag] = struct{}{}
		}
		if len(commandSlc) > 1 {
			recurs(commandSlc[1:], curr, prev)
		}
	default:
		args := strings.Split(v, "=")
		if len(args) > 1 {
			// handle "=" args eg. id=123. args are split by delimiter ','
			if args[0] == "" {
				// todo invalid format eg. var =1 needs to be var=1
				fmt.Println("invalid arg: make sure args are written with no space eg. var=1")
			}
			curr.args[args[0]] = strings.Split(args[1], ",")
			if len(commandSlc) > 1 {
				recurs(commandSlc[1:], curr, prev)
			}
		} else {
			// handle a normal Command eg sched
			if len(commandSlc) > 1 {
				if prev {
					curr.subCommand = newEmptyCommand()
					curr.subCommand.name = commandSlc[0]
					recurs(commandSlc[1:], curr.subCommand, true)
				} else {
					curr.name = commandSlc[0]
					recurs(commandSlc[1:], curr, true)
				}
			} else {
				if prev {
					curr.subCommand = newEmptyCommand()
					curr.subCommand.name = commandSlc[0]
				} else {
					curr.name = commandSlc[0]
				}
			}
		}
	}
}

// ? routing
// ? ========================================================================================================================================================

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
