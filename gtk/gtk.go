package main

import (
	"github.com/conformal/gotk3/cairo"
	"github.com/conformal/gotk3/gtk"
)

func gtkMain() error {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}
	win.SetTitle("smash")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetDefaultSize(640, 480)

	da, err := gtk.DrawingAreaNew()
	check(err)
	da.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
	})
	win.Add(da)

	win.ShowAll()
	gtk.Main()
	return nil
}
