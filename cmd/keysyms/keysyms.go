package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

// go run tools/keysyms.go < /usr/include/1/keysymdef.h | gofmt > xlib/keysym.go

func main() {
	fmt.Printf(`package xlib

//go:generate stringer -type=KeySym
type KeySym int

const (
`)
	s := bufio.NewScanner(os.Stdin)
	re := regexp.MustCompile(`^#define XK_(\S+) *(\S+)`)
	for s.Scan() {
		matches := re.FindStringSubmatch(s.Text())
		if matches != nil {
			name, val := matches[1], matches[2]
			fmt.Printf("\tKey_%s KeySym = %s\n", name, val)
		}
	}
	fmt.Printf(")\n")
	if err := s.Err(); err != nil {
		log.Fatalf("%s", err)
	}
}
