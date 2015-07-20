package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var lines = flag.Int("lines", 1000, "number of lines")
var delay = flag.Int("delay", 0, "delay (in ms) between chars")

func writeLine(i int) {
	line := []byte(fmt.Sprintf("this is line %d\n", i))
	if *delay > 0 {
		for i := 0; i < len(line); i++ {
			os.Stdout.Write(line[i : i+1])
			time.Sleep(time.Duration(*delay) * time.Millisecond)
		}
	} else {
		os.Stdout.Write(line)
	}
}

func main() {
	flag.Parse()
	for i := 0; i < *lines; i++ {
		writeLine(i + 1)
	}
}
