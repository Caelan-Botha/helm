package helm

import (
	"fmt"
	"strings"
)

func ZeroCommand() Command {
	var zeroCommand Command
	return zeroCommand
}

type Command struct {
	raw        []byte
	name       string
	args       map[string]arg
	flags      map[byte]struct{}
	subCommand *Command
}

type arg struct {
	values []string
}

func (a arg) Values() []string {
	return a.values
}

func newCommand(raw []byte, routes map[string]Route) (Command, error) {
	command := Command{
		raw:   raw,
		args:  make(map[string]arg, 0),
		flags: make(map[byte]struct{}),
	}

	// convert CRLF to LF
	text := strings.Replace(string(raw), "\n", "", -1)
	splitBySpaces := strings.Split(strings.Trim(text, " "), " ")

	if len(splitBySpaces) == 0 {
		// todo error no text
		return ZeroCommand(), fmt.Errorf("no command")
	}

	// name of main command
	command.name = splitBySpaces[0]

	// exit early if invalid
	if _, exists := routes[command.name]; !exists {
		return ZeroCommand(), fmt.Errorf("there is no matching route for command name: %s", command.name)
	}

	// rest of the text without name
	commandExcludingName := splitBySpaces[1:]
	// skipNext is used when an arg is detected, since the next word will be the value of the arg
	skipNext := false

	for i := 0; i < len(commandExcludingName); i++ {
		if skipNext {
			// continue if the last iteration was handling an arg
			skipNext = false
			continue
		}
		word := commandExcludingName[i]
		wordChars := []byte(word)
		switch wordChars[0] {
		case '-':
			// handle flags
			trimDashes := strings.Trim(word, "-")
			chars := []byte(trimDashes)
			for _, char := range chars {
				command.flags[char] = struct{}{}
			}
		default:
			// check for matching sub command
			if routes[command.name].subCommands != nil {
				if _, exists := routes[command.name].subCommands[word]; exists {
					rest := strings.Join(commandExcludingName[i:], " ")
					sub, err := newCommand([]byte(rest), routes[command.name].subCommands)
					if err != nil {
						return ZeroCommand(), fmt.Errorf("failed to build sub command: %w", err)
					}
					command.subCommand = &sub
					return command, nil
				}
			}

			// if it isn't a sub command it must be an arg
			skipNext = true
			if len(commandExcludingName) <= i+1 {
				return ZeroCommand(), fmt.Errorf("missing value for argument: %s", commandExcludingName[i])
			}

			args := arg{
				values: make([]string, 0),
			}
			splitArgs := strings.Split(commandExcludingName[i+1], ",")
			for _, argVal := range splitArgs {
				args.values = append(args.values, argVal)
			}

			command.args[commandExcludingName[i]] = args
		}
	}
	return command, nil
}

func (c Command) Name() string {
	return c.name
}

func (c Command) Arg(argName string) []string {
	if arg, exists := c.args[argName]; exists {
		return arg.values
	}
	return nil
}

func (c Command) Args() map[string]arg {
	return c.args
}

func (c Command) Flag(flag byte) bool {
	if _, exists := c.flags[flag]; exists {
		return true
	}
	return false
}

func (c Command) Flags() map[byte]struct{} {
	return c.flags
}
