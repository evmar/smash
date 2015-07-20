package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"smash/base"
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

	term *TermBuf
}

func (win *Window) Mapped() {
	go func() {
		win.term.runBash()
		win.dpy.Quit()
	}()
}

func (w *Window) Draw(cr *cairo.Context) {
	w.term.Draw(cr)
}

func (w *Window) Key(key *base.Key) {
	w.term.Key(key)
}

func (w *Window) Scroll(dy int) {
	w.term.Scroll(dy)
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
	win.term = NewTermBuf(win)

	dpy.Loop(win.xwin)

	// For some reason, things wait a bit on shutdown unless we force-exit.
	if *cpuprofile == "" {
		os.Exit(0)
	}
}
