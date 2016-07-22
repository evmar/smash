package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/martine/smash/bash"
)

func main() {
	b, err := bash.StartBash()
	if err != nil {
		log.Fatalf("start failed: %s", err)
	}

	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("text to complete> ")
		if !s.Scan() {
			break
		}
		_, exps, err := b.Complete(s.Text())
		if err != nil {
			log.Fatalf("run failed: %s", err)
		}
		for _, exp := range exps {
			fmt.Printf("  %q\n", exp)
		}
	}
	if err := s.Err(); err != nil {
		log.Fatalf("scan: %s", err)
	}
}
