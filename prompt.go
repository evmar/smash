package main

import (
	"log"

	"github.com/martine/gocairo/cairo"

	"smash/keys"
	"smash/readline"
	"smash/shell"
	"smash/ui"
	"smash/vt100"
)

type PromptView struct {
	ViewBase
	mf *MonoFont

	shell    *shell.Shell
	readline *readline.ReadLine

	cwin *CompletionWindow
}

type CompletionWindow struct {
	win         ui.Win
	completions []string
}

type CompletionView struct {
}

func NewPromptView(parent View, config *readline.Config, shell *shell.Shell, accept func(string) bool) *PromptView {
	pv := &PromptView{
		ViewBase: ViewBase{Parent: parent},
		mf:       parent.GetWindow().font,
		shell:    shell,
		readline: config.NewReadLine(),
	}
	pv.readline.Accept = accept
	pv.readline.StartComplete = pv.StartComplete
	return pv
}

func (pv *PromptView) Draw(cr *cairo.Context) {
	pv.mf.Use(cr, false)

	cr.MoveTo(0, float64(pv.mf.ch-pv.mf.descent))
	var line []vt100.Cell
	line = append(line, vt100.Cell{Ch: '$'})
	line = append(line, vt100.Cell{Ch: ' '})
	var bold vt100.Attr
	bold.SetBright(true)
	for _, c := range pv.readline.Text {
		line = append(line, vt100.Cell{Ch: rune(c), Attr: bold})
	}
	drawTerminalLine(cr, pv.mf, 0, line)

	if pv.readline.Pos >= 0 {
		ch := rune(0)
		if pv.readline.Pos < len(pv.readline.Text) {
			ch = rune(pv.readline.Text[pv.readline.Pos])
		}
		drawCursor(cr, pv.mf, 0, pv.readline.Pos+2, ch)
	}
}

func (pv *PromptView) Height() int {
	return pv.mf.ch
}

func (pv *PromptView) Key(key keys.Key) {
	if key.Sym == keys.NoSym {
		return
	}
	pv.readline.Key(key)
	pv.Dirty()
}

func (pv *PromptView) Scroll(dy int) {
}

func (pv *PromptView) StartComplete(cb func(string, int), text string, pos int) {
	go func() {
		ofs, completions, err := pv.shell.Complete(text)
		log.Printf("comp %v => %v %v %v", text, ofs, completions, err)
		if len(completions) == 1 {
			// Keep text up to the place where completion started and text
			// after the cursor position.  This is consistent with bash, at
			// least.
			text = text[:ofs] + completions[0] + text[pos:]
			pos = ofs + len(completions[0])
			pv.Enqueue(func() {
				cb(text, pos)
				pv.Dirty()
			})
		} else if len(completions) > 1 {
			pv.Enqueue(func() {
				pv.ShowCompletions(completions)
			})
		}
	}()
}

func (pv *PromptView) ShowCompletions(completions []string) {
	pv.cwin = NewCompletionWindow(pv.GetWindow().ui, completions)
}

func NewCompletionWindow(ui ui.UI, completions []string) *CompletionWindow {
	cwin := &CompletionWindow{
		completions: completions,
	}
	cwin.win = ui.NewWindow(cwin, false)
	cwin.win.SetSize(300, 200)
	cwin.win.SetPosition(300, 200)
	//cwin.view = &CompletionView{}
	cwin.win.Show()
	return cwin
}

func (cw *CompletionWindow) Mapped() {
}
func (cw *CompletionWindow) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 0)
	cr.Paint()
}
func (cw *CompletionWindow) Key(key keys.Key) {
}
func (cw *CompletionWindow) Scroll(dy int) {
}
