package readline

import (
	"log"

	"smash/keys"
)

type Config struct {
	Bindings map[string]string

	History    []string
	HistoryPos int
}

func NewConfig() *Config {
	c := &Config{
		Bindings: DefaultBindings(),
	}

	return c
}

type Complete struct {
	rl        *ReadLine
	cancelled bool
}

func (c *Complete) Results(input string, pos int) {
	if c.cancelled {
		return
	}
	c.rl.finishComplete(input, pos)
}

type ReadLine struct {
	Config *Config

	// User-entered text.  TODO: runes?
	Text []byte

	// Cursor position, or -1 for no cursor.
	Pos int

	Accept          func(string)
	pendingComplete *Complete
	StartComplete   func(c *Complete, input string, pos int)
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

func (rl *ReadLine) Key(key keys.Key) {
	rl.cancelPending()

	bind := rl.Config.Bindings[key.Spec()]
	if bind == "" {
		log.Printf("readline: unhandled key %q", key.Spec())
		return
	}

	cmd := commands[bind]
	if cmd == nil {
		log.Printf("readline: unknown binding %q for key %q", bind, key.Spec())
		return
	}

	cmd(rl, key)
}

func (rl *ReadLine) cancelPending() {
	if rl.pendingComplete == nil {
		return
	}
	rl.pendingComplete.cancelled = true
	rl.pendingComplete = nil
}

func (rl *ReadLine) startComplete() {
	rl.cancelPending()

	pc := &Complete{rl: rl}
	rl.pendingComplete = pc
	rl.StartComplete(pc, rl.String(), rl.Pos)
}

func (rl *ReadLine) finishComplete(text string, pos int) {
	rl.Text = []byte(text)
	rl.Pos = pos
	rl.pendingComplete = nil
}
