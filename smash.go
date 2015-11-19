package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"smash/base"
	"smash/keys"
	"smash/ui"
	"smash/ui/gtk"
	"syscall"

	"github.com/martine/gocairo/cairo"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

const EIO syscall.Errno = 5

var g_anims *base.AnimSet
var gui ui.UI

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Window struct {
	ui  ui.UI
	win ui.Win

	view View
	font *MonoFont
}

type View interface {
	GetWindow() *Window
	Draw(cr *cairo.Context)
	Key(key keys.Key)
	Scroll(dy int)
	Dirty()
	Enqueue(f func())
}

type ViewBase struct {
	Parent View
}

func (vb *ViewBase) GetWindow() *Window {
	return vb.Parent.GetWindow()
}

func (vb *ViewBase) Dirty() {
	vb.Parent.Dirty()
}

func (vb *ViewBase) Enqueue(f func()) {
	vb.Parent.Enqueue(f)
}

func (win *Window) GetWindow() *Window {
	return win
}

func (win *Window) Mapped() {
	panic("x")
}

func (w *Window) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()
	w.view.Draw(cr)
}

func (w *Window) Key(key keys.Key) {
	w.view.Key(key)
}

func (w *Window) Scroll(dy int) {
	w.view.Scroll(dy)
}

func (w *Window) Dirty() {
	w.win.Dirty()
}

func (w *Window) Enqueue(f func()) {
	w.ui.Enqueue(f)
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// ui := xlib.OpenDisplay(anims)
	ui := gtk.Init()
	gui = ui

	win := &Window{
		ui:   ui,
		font: NewMonoFont(),
	}
	win.win = ui.NewWindow(win)
	win.view = NewLogView(win)

	ui.Loop(win.win)

	// For some reason, things wait a bit on shutdown unless we force-exit.
	if *cpuprofile == "" {
		os.Exit(0)
	}
}
