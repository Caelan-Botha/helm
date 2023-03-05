package main

import (
	"github.com/Caelan-Botha/helm/helm"
	"github.com/Caelan-Botha/helm/ui"
)

func main() {
	term := ui.NewUI()
	h := helm.NewHelm(term.Reader, term.Writer)

	h.AddRoute("hello", helm.HelloMainFunc, helm.HelloSubCommandsMap())

	go h.Start()
	term.Start()
}
