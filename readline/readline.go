package readline

import (
	"fmt"
	"log"
	"strings"
)

const (
	Control = uint(1 << iota)
)

type ReadLine struct {
	Text []byte
	Pos  int
}

func NewReadLine() *ReadLine {
	return &ReadLine{}
}

func (rl *ReadLine) String() string {
	return string(rl.Text)
}

func (rl *ReadLine) Insert(ch byte) {
	if rl.Pos == len(rl.Text) {
		rl.Text = append(rl.Text, ch)
	} else {
		rl.Text = append(rl.Text, 0)
		copy(rl.Text[rl.Pos+1:], rl.Text[rl.Pos:])
		rl.Text[rl.Pos] = ch
	}
	rl.Pos++
}

func (rl *ReadLine) Home() {
	rl.Pos = 0
}

func showChar(ch rune) string {
	if ch >= ' ' && ch <= '~' {
		return fmt.Sprintf("%c", ch)
	} else {
		return fmt.Sprintf("%#x", ch)
	}
}

func showMods(mods uint) string {
	var out []string
	if mods&Control != 0 {
		out = append(out, "Control")
	}
	return strings.Join(out, "-")
}

func showKey(key rune, mods uint) string {
	m := showMods(mods)
	if m != "" {
		return m + "-" + showChar(key)
	}
	return showChar(key)
}

func (rl *ReadLine) Key(key rune, mods uint) bool {
	switch {
	case mods == Control && key == 'a':
		rl.Home()
	case mods == 0 && key >= ' ' && key <= '~':
		rl.Insert(byte(key))
	default:
		log.Printf("readline: unhandled key %s", showKey(key, mods))
	}
	return false
}
