package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"smash/ui/gtk"
	"syscall"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

const EIO syscall.Errno = 5

func check(err error) {
	if err != nil {
		panic(err)
	}
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
