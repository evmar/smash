package main

import (
	"log"

	"github.com/martine/gocairo/cairo"

	"smash/keys"
	"smash/readline"
	"smash/shell"
)

type PromptView struct {
	ViewBase
	mf *MonoFont

	shell    *shell.Shell
	readline *readline.ReadLine
}

func NewPromptView(parent View, config *readline.Config, shell *shell.Shell, accept func(string) bool) *PromptView {
	pb := &PromptView{
		ViewBase: ViewBase{Parent: parent},
		mf:       parent.GetWindow().font,
		shell:    shell,
		readline: config.NewReadLine(),
	}
	pb.readline.Accept = accept
	pb.readline.StartComplete = pb.StartComplete
	return pb
}

func (pb *PromptView) Draw(cr *cairo.Context) {
	pb.mf.Use(cr, false)

	cr.MoveTo(0, float64(pb.mf.ch-pb.mf.descent))
	var line []TerminalChar
	line = append(line, TerminalChar{Ch: '$'})
	line = append(line, TerminalChar{Ch: ' '})
	var bold Attr
	bold.SetBright(true)
	for _, c := range pb.readline.Text {
		line = append(line, TerminalChar{Ch: rune(c), Attr: bold})
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
		ofs, completions, err := pb.shell.Complete(text)
		log.Printf("comp %v => %v %v %v", text, ofs, completions, err)
		comp := "foo"
		if len(completions) > 0 {
			comp = completions[0]
		}
		pb.Enqueue(func() {
			c.Results(comp, 0)
			pb.Dirty()
		})
	}()
}
