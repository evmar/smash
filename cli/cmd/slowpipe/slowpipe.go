package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

func run() error {
	r := bufio.NewReader(os.Stdin)
	for {
		c, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err  = os.Stdout.Write([]byte{c})
		if err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "slowpipe: %s", err)
		os.Exit(1)
	}
}