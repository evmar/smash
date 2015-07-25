package readline

import (
	"fmt"
	"log"
)

const (
	Control = uint(1 << iota)
)

type Config struct {
	Bindings map[string]string
}

func NewConfig() *Config {
	c := &Config{
		Bindings: map[string]string{},
	}
	for ch := ' '; ch <= '~'; ch++ {
		c.Bindings[fmt.Sprintf("%c", ch)] = "self-insert"
	}
	c.Bindings["C-a"] = "beginning-of-line"
	return c
}

type ReadLine struct {
	Config *Config
	Text   []byte
	Pos    int
}

type Key struct {
	Ch   rune
	Mods uint
}

func (c *Config) NewReadLine() *ReadLine {
	return &ReadLine{Config: c}
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
		return fmt.Sprintf("\\u%4x", ch)
	}
}

func (k Key) Spec() string {
	spec := ""
	if k.Mods&Control != 0 {
		spec += "C-"
	}
	spec += showChar(k.Ch)
	return spec
}

func (rl *ReadLine) Key(key Key) bool {
	bind := rl.Config.Bindings[key.Spec()]
	if bind == "" {
		log.Printf("readline: unhandled key %q", key.Spec())
		return false
	}

	cmd := commands[bind]
	if cmd == nil {
		log.Printf("readline: unknown binding %q for key %q", bind, key.Spec())
		return false
	}

	cmd(rl, key)
	return false
}
