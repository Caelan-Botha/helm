package main

import (
	"fmt"
	"github.com/Caelan-Botha/helm/helm"
)

func main() {
	//term := ui.NewUI()
	//h := helm.NewHelm(term.Reader, term.Writer)
	h, term := helm.NewHelmUI()

	h.AddRoute("test", testHelm, nil)

	go h.Start()
	term.Start()
}

func testHelm(t *helm.Helm, subCommands helm.SubCommandFuncsMap) error {
	fmt.Println("hi")
	fmt.Println(t.CurrentCommand().Name())
	fmt.Println(t.CurrentCommand().Args())
	fmt.Println(t.CurrentCommand().Flags())
	return nil
}
