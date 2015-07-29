package main

import (
	"smash/base"
	"smash/readline"

	"github.com/martine/gocairo/cairo"
)

type PromptBuf struct {
	ViewBase
	metrics cairo.FontExtents

	rlconfig *readline.Config
	readline *readline.ReadLine
}

func NewPromptBuf(parent View) *PromptBuf {
	config := readline.NewConfig()
	return &PromptBuf{
		ViewBase: ViewBase{Parent: parent},
		rlconfig: config,
		readline: config.NewReadLine(),
	}
}

func (pb *PromptBuf) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()

	cr.SetSourceRGB(0, 0, 0)
	cr.SelectFontFace("monospace", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(14)
	if pb.metrics.MaxXAdvance == 0 {
		cr.FontExtents(&pb.metrics)
	}

	text := "$ "
	if pb.readline != nil {
		text += pb.readline.String()
	}

	cr.Translate(100, 100)
	cr.MoveTo(0, float64(pb.metrics.Height-pb.metrics.Descent))
	cr.ShowText(text)
	pb.Dirty()
}

func (pb *PromptBuf) Key(key *base.Key) {
	pb.readline.Key(readline.Key{Ch: rune(key.Text[0])})
}

func (pb *PromptBuf) Scroll(dy int) {
}
