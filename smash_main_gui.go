// +build !headless

package smash

import "github.com/martine/smash/ui/gtk"

func SmashMain() {
	ui := gtk.Init()
	win := &Window{
		ui:   ui,
		font: NewMonoFont(),
	}
	win.win = ui.NewWindow(win, true)
	// Use the font once to get its metrics.
	cr := win.win.GetCairo()
	win.font.Use(cr, false)
	w, h := win.font.cw*80, win.font.ch*24
	win.win.SetSize(w, h)
	var err error
	win.View, err = NewLogView(win, h)
	if err != nil {
		panic(err)
	}
	win.win.Show()
	ui.Loop()
}
