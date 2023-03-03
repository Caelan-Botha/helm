package helm

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// ? command
// ? ========================================================================================================================================================
// command represents the structure of a commands name with args and flags, as well as a pointer to the sub commands
type command struct {
	name       string
	args       map[string][]string // args are split by delimiter ',' eg. arg=123,555,etc.
	flags      map[byte]struct{}   // this is an efficient way of implementing a set, only use the keys (struct{} allocates 0 memory)
	subCommand *command
}

func newEmptyCommand() *command {
	return &command{
		args:  make(map[string][]string, 0),
		flags: make(map[byte]struct{}, 0),
	}
}

// CommandExecutor
// types for grouping commands and their sub-commands
// the main command will be a command executor that will branch to the sub-commands if needed.
// this is done by registering the pair of commands/sub-commands as a route on the schell
// the main command calls sub-commands via the schell dependency injection
type CommandExecutor func(term *CaeTerm) error
type SubCommandFuncsMap map[string]CommandExecutor

// route represents a command and its related sub commands
// mainCommandFunc: is the function that is run when the command is called, this should contain the branching logic for sub commands eg. a switch statement
type route struct {
	mainCommandFunc    CommandExecutor
	subCommandFuncsMap map[string]CommandExecutor
}

func newRoute(mainCommandFunc CommandExecutor, subCommandFuncsMap SubCommandFuncsMap) route {
	return route{
		mainCommandFunc:    mainCommandFunc,
		subCommandFuncsMap: subCommandFuncsMap,
	}
}

// ? caeTerm
// ? ========================================================================================================================================================

type CaeTerm struct {
	in  io.Reader
	out io.Writer

	routes map[string]route

	currentCmd *command

	quitCh chan struct{}
}

func NewTerm() *CaeTerm {
	return &CaeTerm{
		in:     os.Stdin,
		out:    os.Stdout,
		routes: make(map[string]route),
		quitCh: make(chan struct{}),
	}
}

func (t *CaeTerm) AddRoute(name string, mainCommandFunc CommandExecutor, subCommandFuncsMap SubCommandFuncsMap) error {
	if _, exists := t.routes[name]; exists {
		return fmt.Errorf("a route with name: %s already exists", name)
	}

	t.routes[name] = newRoute(mainCommandFunc, subCommandFuncsMap)

	return nil
}

func (t *CaeTerm) Start() {

	go t.inputLoop()

	<-t.quitCh
}

// ? input
// ? ========================================================================================================================================================

func (t *CaeTerm) inputLoop() {
	parseDoneCh := make(chan struct{})
	errCh := make(chan error)
	for {
		reader := bufio.NewReader(t.in)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("read line: %s-\n", line)
		t.parseCommand(line, parseDoneCh)
		<-parseDoneCh
		t.routeCommand(errCh)
	}
}

// ? parsing
// ? ========================================================================================================================================================
// handle input from command line
func (t *CaeTerm) parseCommand(text string, parseDoneCh chan struct{}) {
	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)
	commandSlc := strings.Split(strings.Trim(text, " "), " ")
	// reset current command
	t.currentCmd = newEmptyCommand()
	// recursively build command
	recurs(commandSlc, t.currentCmd, false)
	parseDoneCh <- struct{}{}
}

// recursively builds the command struct passed in as a pointer
func recurs(commandSlc []string, curr *command, prev bool) {
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
			// handle a normal command eg sched
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

// route the command to the appropriate handler
func (t *CaeTerm) routeCommand(errCh chan error) {
	if _, exists := t.routes[t.currentCmd.name]; !exists {
		// command isn't registered
		errCh <- fmt.Errorf("unknow command: %s", t.currentCmd.name)
		return
	}
	// execute command main func which will handle any sub commands
	err := t.routes[t.currentCmd.name].mainCommandFunc(t)
	if err != nil {
		// todo handle error
		errCh <- err
	}
}
