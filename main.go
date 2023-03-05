package main

import (
	"github.com/Caelan-Botha/helm/helm"
)

func main() {
	//term := ui.NewUI()
	//h := helm.NewHelm(term.Reader, term.Writer)
	h, term := helm.NewHelmUI()

	h.AddRoute("hello", helm.HelloMainFunc, helm.HelloSubCommandsMap())

	go h.Start()
	term.Start()
}
