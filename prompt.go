package main

import (
	"smash/keys"
	"smash/readline"
	"time"

	"github.com/martine/gocairo/cairo"
)

type PromptBuf struct {
	ViewBase
	mf *MonoFont

	rlconfig *readline.Config
	readline *readline.ReadLine
}

func NewPromptBuf(parent View) *PromptBuf {
	config := readline.NewConfig()
	pb := &PromptBuf{
		ViewBase: ViewBase{Parent: parent},
		mf:       GetMonoFont(),
		rlconfig: config,
	}
	pb.Reset()
	return pb
}

func (pb *PromptBuf) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()

	cr.SetSourceRGB(0, 0, 0)
	pb.mf.Use(cr)

	cr.MoveTo(0, float64(pb.mf.ch-pb.mf.descent))
	var line []TerminalChar
	line = append(line, TerminalChar{Ch: '$'})
	line = append(line, TerminalChar{Ch: ' '})
	for _, c := range pb.readline.Text {
		line = append(line, TerminalChar{Ch: rune(c)})
	}
	drawTerminalLine(cr, pb.mf, 0, line)
	ch := rune(0)
	if pb.readline.Pos < len(pb.readline.Text) {
		ch = rune(pb.readline.Text[pb.readline.Pos])
	}
	drawCursor(cr, pb.mf, 0, pb.readline.Pos+2, ch)
}

func (pb *PromptBuf) Key(key keys.Key) {
	if key.Sym == keys.NoSym {
		return
	}
	pb.readline.Key(key)
	pb.Dirty()
}

func (pb *PromptBuf) Scroll(dy int) {
}

func (pb *PromptBuf) StartComplete(c *readline.Complete, text string, pos int) {
	go func() {
		time.Sleep(500 * time.Millisecond)
		pb.Enqueue(func() {
			c.Results("foo", 0)
			pb.Dirty()
		})
	}()
}

func (pb *PromptBuf) Reset() {
	pb.readline = pb.rlconfig.NewReadLine()
	pb.readline.Accept = pb.Reset
	pb.readline.StartComplete = pb.StartComplete
}
