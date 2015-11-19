package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"smash/ui/gtk"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

	ui := gtk.Init()
	win := &Window{
		ui:   ui,
		font: NewMonoFont(),
	}
	win.win = ui.NewWindow(win)
	win.view = NewLogView(win)
	ui.Loop()

	// For some reason, things wait a bit on shutdown unless we force-exit.
	if *cpuprofile == "" {
		os.Exit(0)
	}
}
