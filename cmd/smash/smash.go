package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/evmar/smash"
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

	smash.SmashMain()

	// For some reason, things wait a bit on shutdown.
	// Maybe some sort of finalizers running?
}
