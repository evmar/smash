package main

import (
	"os/exec"

	"github.com/martine/smash"
	"github.com/martine/smash/ui/gtk"
)

func main() {
	ui := gtk.Init()

	win := smash.NewWindow(ui)

	term := smash.NewTermView(win)
	win.View = term

	w, h := term.Measure(24, 80)
	win.GetUiWindow().SetSize(w, h)

	term.Start(exec.Command("bash"))

	win.Show()
	ui.Loop()
}
