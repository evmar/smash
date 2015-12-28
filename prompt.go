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

	start, end  int
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
	bind := pv.readline.Key(key)
	if pv.cwin != nil {
		if bind == "self-insert" {
			pv.UseCompletions(pv.cwin.start, pv.readline.Pos, pv.cwin.completions, false)
		} else {
			pv.cwin.Close()
			pv.cwin = nil
		}
	}

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
	// log.Printf("filterPrefix %q %q", text, completions)

	// First filter completions to those with the prefix.
	if len(text) > 0 {
		comps := []string{}
		for _, c := range completions {
			if strings.HasPrefix(c, text) {
				comps = append(comps, c)
			}
		}
		if len(comps) > 0 {
			completions = comps
		}
	}

	if len(completions) == 0 {
		return text, []string{}
	}

	// Then find the longest common prefix of those completions.
	// (Consider input "l", completions [log logview].  We want to
	// expand to the shared prefix "log" despite there still being
	// multiple completions available.)
	for i := 0; ; i++ {
		for _, comp := range completions {
			if i == len(comp) || comp[i] != completions[0][i] {
				return comp[:i], completions
			}
		}
	}
}

func (pv *PromptView) StartComplete() {
	text := string(pv.readline.Text)
	end := pv.readline.Pos
	go func() {
		start, completions, err := pv.shell.Complete(text)

		// Consider the input:
		//   ls l<tab>
		// text: "ls l"
		// end, the cursor postion: 4 (after the "l")
		// start, the completion beginning: 3 (before the "l")
		// completions: [log logview ...]
		log.Printf("comp %v %v => %v %v %v", text, start, end, completions, err)
		pv.Enqueue(func() {
			pv.UseCompletions(start, end, completions, true)
		})
	}()
}

func (pv *PromptView) UseCompletions(start, end int, completions []string, expand bool) {
	text := string(pv.readline.Text)

	newText, completions := filterPrefix(text[start:end], completions)
	log.Printf("usecomp %q %q %q", expand, newText, completions)

	if expand {
		text = text[:start] + newText + text[end:]
		end = start + len(newText)
		pv.readline.Text = []byte(text)
		pv.readline.Pos = end
		pv.Dirty()
	}

	if len(completions) > 1 || !expand {
		pv.ShowCompletions(start, end, completions)
	}
}

func (pv *PromptView) ShowCompletions(start, end int, completions []string) {
	x, y := pv.delegate.GetPromptAbsolutePosition(pv)
	x += (len("$ ") + start) * pv.mf.cw
	y -= pv.mf.ch

	if pv.cwin != nil {
		pv.cwin.Close()
	}
	pv.cwin = NewCompletionWindow(pv, x, y, start, end, completions)
}

func NewCompletionWindow(pv *PromptView, x, y int, start, end int, completions []string) *CompletionWindow {
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
		start:       start,
		end:         end,
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

		prefixLen := cw.end - cw.start
		cw.font.Use(cr, true)
		cr.ShowText(c[:prefixLen])
		cw.font.Use(cr, false)
		cr.ShowText(c[prefixLen:])
	}
}

func (cw *CompletionWindow) cycle(delta int) {
	cw.sel = ((cw.sel + delta) + len(cw.completions)) % len(cw.completions)
	cw.win.Dirty()
}

func (cw *CompletionWindow) Key(key keys.Key) bool {
	switch key.Spec() {
	case "Tab":
		if len(cw.completions) == 1 {
			cw.pv.OnCompletion(cw.completions[0][cw.start:])
		} else {
			cw.cycle(1)
		}
		return true
	case "Down", "C-n":
		cw.cycle(1)
		return true
	case "Up":
		cw.cycle(-1)
		return true
	case "Enter":
		cw.pv.OnCompletion(cw.completions[cw.sel][cw.start:])
		return true
	case "Esc":
		cw.pv.OnCompletion("")
		return true
	default:
		log.Printf("CompletionWindow unhandled: %s", key)
		return false
	}
}

func (cw *CompletionWindow) Scroll(dy int) {
	panic("x")
}

func (cw *CompletionWindow) Close() {
	cw.win.Close()
}
