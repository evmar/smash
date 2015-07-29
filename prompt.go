package main

import (
	"smash/keys"
	"smash/readline"

	"github.com/martine/gocairo/cairo"
)

type PromptBuf struct {
	ViewBase
	metrics Metrics

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
	if pb.metrics.cw == 0 {
		pb.metrics.FillFromCairo(cr)
	}

	cr.MoveTo(0, float64(pb.metrics.ch-pb.metrics.descent))
	var line []TerminalChar
	line = append(line, TerminalChar{Ch: '$'})
	line = append(line, TerminalChar{Ch: ' '})
	for _, c := range pb.readline.Text {
		line = append(line, TerminalChar{Ch: rune(c)})
	}
	drawTerminalLine(cr, &pb.metrics, 0, line)
	ch := rune(0)
	if pb.readline.Pos < len(pb.readline.Text) {
		ch = rune(pb.readline.Text[pb.readline.Pos])
	}
	drawCursor(cr, &pb.metrics, 0, pb.readline.Pos+2, ch)
	pb.Dirty()
}

func (pb *PromptBuf) Key(key keys.Key) {
	if key.Sym == keys.NoSym {
		return
	}
	pb.readline.Key(key)
}

func (pb *PromptBuf) Scroll(dy int) {
}
