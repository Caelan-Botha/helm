package helm

import (
	"fmt"
	"github.com/Caelan-Botha/helm/ui/formatters"
	"strings"
)

type ErrMissingArg struct {
	ArgName string
}

func (e ErrMissingArg) Error() string {
	return fmt.Sprintf("missing required arg: %s", e.ArgName)
}

func StartExample() {
	//h := NewHelm(os.Stdin, os.Stdout)
	h, u := NewHelmUI()
	h.RegisterRoute("hello", Route{
		MainCommand: HelloMainFunc,
		SubCommands: HelloSubCommands(),
	})

	h.RegisterRoute("json", Route{
		MainCommand: PrintJSONMainFunc,
		SubCommands: nil,
	})

	go h.Start()
	u.Start()
}

func HelloMainFunc(h *Helm) error {
	// check flags
	str := "Hi there!"
	if len(h.CurrentCommand().Flags()) > 0 {
		str += " flag: "
		flags := make([]string, 0)
		for f := range h.currentCmd.Flags() {
			flags = append(flags, string(f))
		}
		str += strings.Join(flags, ",")
	}
	// check args
	if len(h.CurrentCommand().Args()) > 0 {
		str += " args: "
		args := make([]string, 0)
		for argName, argValue := range h.currentCmd.Args() {
			argVals := make([]string, 0)
			for _, arg := range argValue.Values() {
				argVals = append(argVals, arg)
			}
			args = append(args, argName+"="+strings.Join(argVals, ","))
		}
		str += strings.Join(args, ", ")
	}

	h.OutputString(str)
	return nil
}

func HelloSubCommands() map[string]Route {
	return map[string]Route{
		"one": {
			MainCommand: func(helm *Helm) error {
				helm.OutputString("running sub-command: one")
				return nil
			},
			SubCommands: map[string]Route{
				"uno": {
					MainCommand: func(helm *Helm) error {
						helm.OutputString("running sub-sub-command: uno")
						return nil
					},
					SubCommands: nil,
				},
			},
		},
		"two": {
			MainCommand: func(helm *Helm) error {
				helm.OutputString("running sub-command: two")
				return nil
			},
			SubCommands: nil,
		},
		"three": {
			MainCommand: func(helm *Helm) error {
				helm.OutputString("running sub-command: three")
				return nil
			},
			SubCommands: nil,
		},
	}
}

func PrintJSONMainFunc(h *Helm) error {
	jsonBytes := `{
  "product": "Live JSON generator",
  "version": 3.1,
  "releaseDate": "2014-06-25T00:00:00.000Z",
  "demo": true,
  "person": {
    "id": 12345,
    "name": "John Doe",
    "phones": {
      "home": "800-123-4567",
      "mobile": "877-123-1234"
    },
    "email": [
      "jd@example.com",
      "jd@example.org"
    ],
    "dateOfBirth": "1980-01-02T00:00:00.000Z",
    "registered": true,
    "emergencyContacts": [
      {
        "name": "Jane Doe",
        "phone": "888-555-1212",
        "relationship": "spouse"
      },
      {
        "name": "Justin Doe",
        "phone": "877-123-1212",
        "relationship": "parent"
      }
    ]
  }
}`
	err := h.OutputFormatted([]byte(jsonBytes), formatters.JSONTestFormatter{})
	if err != nil {
		fmt.Println("Error printing formatted", err)
		return err
	}
	return nil
}

//// todo make a better way of getting sub-commands
//// at the moment you have to chain .SubCommand over and over
//// also at the moment there isn't an elegant way to run multiple sub commands (for now you must code layers of ifs) maybe think about doing it recursively
//
//func HelloMainFunc(t *Helm, subCommandsMap SubCommands) error {
//	// check if there was a sub command passed in the command line
//	if t.CurrentCommand().HasSubCommand() {
//		// check if that sub command has been registered on the route
//		if subCommandFunc, exists := subCommandsMap[t.CurrentCommand().SubCommand().Name()]; exists {
//			err := subCommandFunc(t, subCommandsMap)
//			if err != nil {
//				return err
//			}
//		}
//
//		// exit after the sub command (usually because the main command is just the 'help' text)
//		return nil
//	}
//
//	// main functionality eg. no sub command passed, this is nice for providing 'help' info. you can simply print the 'help' info here
//	helloStr := "Hello"
//
//	// check for 's' (short) flag
//	if t.CurrentCommand().Flag('s') {
//		helloStr = "Hi"
//	}
//
//	// check for 'name' arg
//	if listOfNames := t.CurrentCommand().Arg("names"); listOfNames != nil {
//		helloStr += strings.Join(listOfNames, ", ")
//	}
//
//	// write to out
//	_, err := t.out.Write([]byte(helloStr))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return nil
//}
//
//func HelloSubCommandsMap() map[string]CommandFunc {
//	m := make(map[string]CommandFunc)
//
//	m["from"] = func(t *Helm, subCommandsMap SubCommands) error {
//		// handle 'from' sub command
//		helloStr := "Hello"
//
//		// check for 's' (short) flag
//		if t.CurrentCommand().SubCommand().Flag('s') {
//			helloStr = "Hey"
//		}
//
//		// add 'from' to the string
//		helloStr += " from "
//
//		// check for 'name' arg
//		if listOfNames := t.CurrentCommand().SubCommand().Arg("names"); listOfNames != nil {
//			helloStr += strings.Join(listOfNames, ", ")
//		} else {
//			return ErrMissingArg{ArgName: "names"}
//		}
//
//		// write to out
//		_, err := t.out.Write([]byte(helloStr))
//		if err != nil {
//			log.Fatal(err)
//		}
//		return nil
//	}
//
//	m["gunga"] = func(t *Helm, SubCommands SubCommands) error {
//		fmt.Println("saying gunga!")
//		return nil
//	}
//	m["ginga"] = func(t *Helm, SubCommands SubCommands) error {
//		fmt.Println("saying GINGAAAA")
//		return nil
//	}
//
//	return m
//}
