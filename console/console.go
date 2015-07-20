package main

import (
	"fmt"
	"log"
	"os"

	termbox "github.com/nsf/termbox-go"
)

var out = "hi"

func draw() {
	check(termbox.Clear(termbox.ColorDefault, termbox.ColorDefault))
	for i, ch := range out {
		termbox.SetCell(i, 0, ch, termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.Flush()
}

func consoleMain() {
	f, err := os.Create("log")
	check(err)
	lg := log.New(f, "", log.Lmicroseconds)

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	lg.Printf("init")

	draw()
	lg.Printf("draw")

	quit := false
	for !quit {
		lg.Printf("waiting for event")
		ev := termbox.PollEvent()
		lg.Printf("got event %r", ev)
		switch ev.Type {
		case termbox.EventKey:
			if ev.Ch == 'q' {
				quit = true
				break
			}
			out += fmt.Sprintf("%c", ev.Ch)
			break
		case termbox.EventError:
			panic(ev.Err)
		default:
			quit = true
			break
		}
		draw()
	}
}
