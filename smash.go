package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime/pprof"
	"smash/base"
	"smash/keys"
	"smash/ui"
	"smash/ui/xlib"
	"syscall"

	"github.com/martine/gocairo/cairo"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

const EIO syscall.Errno = 5

var anims *base.AnimSet

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Window struct {
	ui  ui.UI
	win ui.Win

	view View
	term *TermBuf
}

type View interface {
	Draw(cr *cairo.Context)
	Key(key keys.Key)
	Scroll(dy int)
	Dirty()
	Enqueue(f func())
}

type ViewBase struct {
	Parent View
}

func (vb *ViewBase) Dirty() {
	vb.Parent.Dirty()
}

func (vb *ViewBase) Enqueue(f func()) {
	vb.Parent.Enqueue(f)
}

func (win *Window) Mapped() {
	if win.term != nil {
		go func() {
			win.term.runCommand(exec.Command("bash"))
			win.ui.Quit()
		}()
	}
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

	anims = base.NewAnimSet()
	ui := xlib.OpenDisplay(anims)

	win := &Window{ui: ui}
	win.win = ui.NewWindow(win)
	if false {
		win.term = NewTermBuf(win)
		win.view = win.term
	} else {
		win.view = NewLogView(win)
	}

	ui.Loop(win.win)

	// For some reason, things wait a bit on shutdown unless we force-exit.
	if *cpuprofile == "" {
		os.Exit(0)
	}
}
