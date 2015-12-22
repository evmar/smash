package main

import (
	"log"

	"github.com/martine/gocairo/cairo"

	"smash/keys"
	"smash/readline"
	"smash/shell"
	"smash/ui"
	"smash/vt100"
	"strings"
)

type PromptDelegate interface {
	OnPromptAccept(string) bool
	GetPromptAbsolutePosition(pv *PromptView) (int, int)
}

type PromptView struct {
	ViewBase
	delegate PromptDelegate

	mf *MonoFont

	shell    *shell.Shell
	readline *readline.ReadLine

	cwin *CompletionWindow
}

type CompletionWindow struct {
	pv  *PromptView
	win ui.Win

	font *MonoFont

	width, height int

	prefix      string
	completions []string
	sel         int
}

func NewPromptView(parent View, delegate PromptDelegate, config *readline.Config, shell *shell.Shell) *PromptView {
	pv := &PromptView{
		ViewBase: ViewBase{Parent: parent},
		delegate: delegate,
		mf:       parent.GetWindow().font,
		shell:    shell,
		readline: config.NewReadLine(),
	}
	pv.readline.Accept = delegate.OnPromptAccept
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

func (pv *PromptView) Key(key keys.Key) bool {
	if key.Sym == keys.NoSym {
		return false
	}
	if pv.cwin != nil {
		if pv.cwin.Key(key) {
			return true
		}
	}
	pv.readline.Key(key)
	pv.Dirty()
	return true
}

func (pv *PromptView) OnCompletion(text string) {
	if len(text) > 0 {
		pv.readline.Text = append(pv.readline.Text, []byte(text)...)
		pv.readline.Pos += len(text)
		pv.Dirty()
	}
	pv.cwin.Close()
	pv.cwin = nil
}

func (pv *PromptView) Scroll(dy int) {
}

// filterPrefix takes some prefix text and a list of completions, and
// computes the longest common prefix of all completions that starts
// with the prefix, as well as all the completions with that prefix.
func filterPrefix(text string, completions []string) (string, []string) {
	log.Printf("filter %v %v", text, completions)

	// First filter completions to those with the prefix.
	if len(text) > 0 {
		comps := []string{}
		for _, c := range completions {
			if strings.HasPrefix(c, text) {
				comps = append(comps, c)
			}
		}
		completions = comps
	}

	// Then find the longest common prefix of those completions.
	for i := 0; ; i++ {
		for _, comp := range completions {
			if i == len(comp) || comp[i] != completions[0][i] {
				return comp[:i], completions
			}
		}
	}
}

func (pv *PromptView) StartComplete(cb func(string, int), text string, pos int) {
	go func() {
		ofs, completions, err := pv.shell.Complete(text)

		// Consider the input:
		//   ls l<tab>
		// text: "ls l"
		// pos, the cursor postion: 4 (after the "l")
		// ofs, the completion beginning: 3 (before the "l")
		// completions: [log logview ...]
		log.Printf("comp %v %v => %v %v %v", text, pos, ofs, completions, err)

		before := text[:ofs]
		after := text[pos:]

		text, completions = filterPrefix(text[ofs:pos], completions)

		pos = ofs + len(text)
		text = before + text + after

		pv.Enqueue(func() {
			cb(text, pos)
			pv.Dirty()
			if len(completions) > 1 {
				pv.ShowCompletions(ofs, completions)
			}
		})
	}()
}

func (pv *PromptView) ShowCompletions(ofs int, completions []string) {
	x, y := pv.delegate.GetPromptAbsolutePosition(pv)
	x += (len("$ ") + ofs) * pv.mf.cw
	y -= pv.mf.ch

	pv.cwin = NewCompletionWindow(pv, x, y, string(pv.readline.Text[ofs:]), completions)
}

func NewCompletionWindow(pv *PromptView, x, y int, prefix string, completions []string) *CompletionWindow {
	w := 0
	for _, c := range completions {
		if len(c) > w {
			w = len(c)
		}
	}
	win := pv.GetWindow()
	cwin := &CompletionWindow{
		pv:          pv,
		font:        win.font,
		prefix:      prefix,
		completions: completions,
		width:       w * win.font.cw,
		height:      len(completions) * win.font.ch,
	}
	cwin.win = win.ui.NewWindow(cwin, false)
	cwin.win.SetSize(cwin.width+2, cwin.height+2)
	cwin.win.SetPosition(x-1, y-1-cwin.height)
	cwin.win.Show()
	return cwin
}

func (cw *CompletionWindow) Update(prefix string) {
}

func (cw *CompletionWindow) Mapped() {
}

func (cw *CompletionWindow) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()

	// Border around the window.
	cr.SetSourceRGB(0, 0, 0)
	cr.Rectangle(0, 0, float64(cw.width+2), float64(cw.height+2))
	cr.Stroke()

	cr.Translate(1, 1)
	y := 0
	for i, c := range cw.completions {
		if i == cw.sel {
			cr.SetSourceRGB(0.90, 0.90, 1)
			cr.Rectangle(0, float64(y), float64(cw.width), float64(cw.font.ch))
			cr.Fill()
		}
		cr.SetSourceRGB(0, 0, 0)
		y += cw.font.ch
		cr.MoveTo(0, float64(y-cw.font.descent))
		cw.font.Use(cr, true)
		cr.ShowText(cw.prefix)
		cw.font.Use(cr, false)
		cr.ShowText(c[len(cw.prefix):])
	}
}

func (cw *CompletionWindow) Key(key keys.Key) bool {
	switch key.Sym {
	case keys.Tab, keys.Down:
		cw.sel = (cw.sel + 1) % len(cw.completions)
		cw.win.Dirty()
		return true
	case keys.Up:
		cw.sel = (cw.sel - 1 + len(cw.completions)) % len(cw.completions)
		cw.win.Dirty()
		return true
	case keys.Enter:
		cw.pv.OnCompletion(cw.completions[cw.sel][len(cw.prefix):])
		return true
	case keys.Esc:
		cw.pv.OnCompletion("")
		return true
	default:
		log.Printf("k %#v", key)
		return false
	}
}

func (cw *CompletionWindow) Scroll(dy int) {
	panic("x")
}

func (cw *CompletionWindow) Close() {
	cw.win.Close()
}
