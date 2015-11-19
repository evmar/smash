package main

import (
	"smash/keys"
	"smash/readline"
	"time"

	"github.com/martine/gocairo/cairo"
)

type PromptView struct {
	ViewBase
	mf *MonoFont

	rlconfig *readline.Config
	readline *readline.ReadLine
	Accept   func(string) bool
}

func NewPromptView(parent View, config *readline.Config, accept func(string) bool) *PromptView {
	pb := &PromptView{
		ViewBase: ViewBase{Parent: parent},
		mf:       parent.GetWindow().font,
		rlconfig: config,
		Accept:   accept,
	}
	pb.Reset()
	return pb
}

func (pb *PromptView) Draw(cr *cairo.Context) {
	pb.mf.Use(cr)

	cr.MoveTo(0, float64(pb.mf.ch-pb.mf.descent))
	var line []TerminalChar
	line = append(line, TerminalChar{Ch: '$'})
	line = append(line, TerminalChar{Ch: ' '})
	for _, c := range pb.readline.Text {
		line = append(line, TerminalChar{Ch: rune(c)})
	}
	drawTerminalLine(cr, pb.mf, 0, line)

	if pb.readline.Pos >= 0 {
		ch := rune(0)
		if pb.readline.Pos < len(pb.readline.Text) {
			ch = rune(pb.readline.Text[pb.readline.Pos])
		}
		drawCursor(cr, pb.mf, 0, pb.readline.Pos+2, ch)
	}
}

func (pb *PromptView) Height() int {
	return pb.mf.ch
}

func (pb *PromptView) Key(key keys.Key) {
	if key.Sym == keys.NoSym {
		return
	}
	pb.readline.Key(key)
	pb.Dirty()
}

func (pb *PromptView) Scroll(dy int) {
}

func (pb *PromptView) StartComplete(c *readline.Complete, text string, pos int) {
	go func() {
		time.Sleep(500 * time.Millisecond)
		pb.Enqueue(func() {
			c.Results("foo", 0)
			pb.Dirty()
		})
	}()
}

func (pb *PromptView) Reset() {
	pb.readline = pb.rlconfig.NewReadLine()
	pb.readline.Accept = pb.Accept
	pb.readline.StartComplete = pb.StartComplete
}
