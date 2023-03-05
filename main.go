package main

import (
	"helm/helm"
	"helm/ui"
)

func main() {
	term := ui.NewUI()
	h := helm.NewHelm(term.Reader, term.Writer)

	h.AddRoute("hello", helm.HelloMainFunc, helm.HelloSubCommandsMap())

	go h.Start()
	term.Start()
}
