package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/evmar/smash/vt100"
)

func main() {
	term := vt100.NewTerminal()
	tr := vt100.NewTermReader(func(f func(t *vt100.Terminal)) {
		f(term)
	})

	r := bufio.NewReader(os.Stdin)
	var err error
	for err == nil {
		err = tr.Read(r)
	}
	if err != nil && err != io.EOF {
		log.Println("ERROR:", err)
		os.Exit(1)
	}
	fmt.Println(term.ToString())
}
