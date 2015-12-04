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

type ReadLine struct {
	Config *Config

	// User-entered text.  TODO: runes?
	Text []byte

	// Cursor position, or -1 for no cursor.
	Pos int

	Accept          func(string) bool
	StartComplete   func(cb func(string, int), input string, pos int)
	pendingComplete bool
	completeId      int
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
	rl.pendingComplete = false

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

func (rl *ReadLine) startComplete() {
	rl.pendingComplete = true
	rl.completeId++
	id := rl.completeId
	rl.StartComplete(func(text string, pos int) {
		if !rl.pendingComplete || id != rl.completeId {
			// cancelled
			return
		}
		rl.finishComplete(text, pos)
	}, rl.String(), rl.Pos)
}

func (rl *ReadLine) finishComplete(text string, pos int) {
	rl.pendingComplete = false
	rl.Text = []byte(text)
	rl.Pos = pos
}
