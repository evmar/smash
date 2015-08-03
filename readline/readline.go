package readline

import (
	"log"

	"smash/keys"
)

type Config struct {
	Bindings map[string]string
}

func NewConfig() *Config {
	c := &Config{
		Bindings: DefaultBindings(),
	}

	return c
}

type pendingComplete struct {
	cancelled bool
}

type ReadLine struct {
	Config          *Config
	Text            []byte
	Pos             int
	Accept          func()
	Enqueue         func(func())
	pendingComplete *pendingComplete
	Complete        func(input string, pos int) (string, int)
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
	rl.pendingComplete = nil

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
	pc := &pendingComplete{}
	go func() {
		newText, newPos := rl.Complete(rl.String(), rl.Pos)
		rl.Enqueue(func() {
			if rl.pendingComplete == pc {
				rl.finishComplete(newText, newPos)
			}
		})
	}()
}

func (rl *ReadLine) finishComplete(text string, pos int) {
	rl.Text = []byte(text)
	rl.Pos = pos
	rl.pendingComplete = nil
}
