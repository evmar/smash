package main

import (
	"os/exec"

	"github.com/martine/smash"
	"github.com/martine/smash/ui/gtk"
)

func main() {
	ui := gtk.Init()

	win := smash.NewWindow(ui)
	// Use the font once to get its metrics.
	// cr := win.GetCairo()
	// win.font.Use(cr, false)
	// w, h := win.font.cw*80, win.font.ch*24
	// win.win.SetSize(w, h)

	term := smash.NewTermView(win)
	term.Start(exec.Command("bash"))
	win.View = term

	win.Show()
	ui.Loop()
}
