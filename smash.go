package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"smash/base"
	"smash/keys"
	"smash/xlib"
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
	dpy  *xlib.Display
	xwin *xlib.Window

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
			win.term.runBash()
			win.dpy.Quit()
		}()
	}
}

func (w *Window) Draw(cr *cairo.Context) {
	w.view.Draw(cr)
}

func (w *Window) Key(key keys.Key) {
	w.view.Key(key)
}

func (w *Window) Scroll(dy int) {
	w.view.Scroll(dy)
}

func (w *Window) Dirty() {
	w.xwin.Dirty()
}

func (w *Window) Enqueue(f func()) {
	w.dpy.Enqueue(f)
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
	dpy := xlib.OpenDisplay(anims)

	win := &Window{dpy: dpy}
	win.xwin = dpy.NewWindow(win)
	if false {
		win.term = NewTermBuf(win)
		win.view = win.term
	} else {
		win.view = NewPromptBuf(win)
	}

	dpy.Loop(win.xwin)

	// For some reason, things wait a bit on shutdown unless we force-exit.
	if *cpuprofile == "" {
		os.Exit(0)
	}
}
