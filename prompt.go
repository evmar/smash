package main

import (
	"smash/keys"
	"smash/readline"
	"time"

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
	pb := &PromptBuf{
		ViewBase: ViewBase{Parent: parent},
		rlconfig: config,
	}
	pb.Reset()
	return pb
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

func (pb *PromptBuf) Complete(text string, pos int) (string, int) {
	time.Sleep(500 * time.Millisecond)
	return "foo", 0
}

func (pb *PromptBuf) Reset() {
	pb.readline = pb.rlconfig.NewReadLine()
	pb.readline.Accept = pb.Reset
	pb.readline.Enqueue = pb.Enqueue
	pb.readline.Complete = pb.Complete
}
