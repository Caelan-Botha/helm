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
	raw   []byte
	name  string
	args  map[string]arg
	flags map[byte]struct{}
}

type arg struct {
	values []string
}

func newCommand(raw []byte) (Command, error) {
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
		return command, fmt.Errorf("no command")
	}

	command.name = splitBySpaces[0]

	commandExcludingName := splitBySpaces[1:]
	skipNext := false

	for i := 0; i < len(commandExcludingName); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		word := commandExcludingName[i]
		wordChars := []byte(word)
		switch wordChars[0] {
		case '-':
			trimDashes := strings.Trim(word, "-")
			chars := []byte(trimDashes)
			for _, char := range chars {
				command.flags[char] = struct{}{}
			}
		default:
			skipNext = true
			if len(commandExcludingName) <= i+1 {
				return command, fmt.Errorf("missing value for argument: %s", commandExcludingName[i])
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
