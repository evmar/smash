package smash

import (
	"log"

	"github.com/evmar/gocairo/cairo"

	"strings"

	"github.com/evmar/smash/keys"
	"github.com/evmar/smash/readline"
	"github.com/evmar/smash/shell"
	"github.com/evmar/smash/ui"
	"github.com/evmar/smash/vt100"
)

type PromptDelegate interface {
	OnPromptAccept(string) bool
	GetPromptAbsolutePosition(pv *PromptView) (int, int)
}

type PromptView struct {
	ViewBase
	delegate PromptDelegate

	font *Font

	shell    *shell.Shell
	readline *readline.ReadLine

	marker PromptMarker
	cwin   *CompletionWindow
}

type PromptMarker struct {
	Width, Height int
}

type CompletionWindow struct {
	pv  *PromptView
	win ui.Win

	font *Font

	width, height int

	start, end  int
	completions []string
	sel         int
}

func NewPromptView(parent View, delegate PromptDelegate, config *readline.Config, shell *shell.Shell) *PromptView {
	font := &Font{
		Name: "sans",
		Size: 16,
	}
	cr := parent.GetWindow().win.GetCairo()
	font.Use(cr, false)

	pv := &PromptView{
		ViewBase: ViewBase{Parent: parent},
		delegate: delegate,
		font:     font,
		shell:    shell,
		marker: PromptMarker{
			Width:  20,
			Height: font.ch - font.descent,
		},
		readline: config.NewReadLine(),
	}
	pv.readline.Accept = delegate.OnPromptAccept
	pv.readline.StartComplete = pv.StartComplete
	return pv
}

func (pv *PromptView) DrawMono(cr *cairo.Context) {
	pv.font.Use(cr, false)

	cr.MoveTo(0, float64(pv.font.ch-pv.font.descent))
	var line []vt100.Cell
	line = append(line, vt100.Cell{Ch: '$'})
	line = append(line, vt100.Cell{Ch: ' '})
	var bold vt100.Attr
	bold.SetBright(true)
	for _, c := range pv.readline.Text {
		line = append(line, vt100.Cell{Ch: rune(c), Attr: bold})
	}
	drawTerminalLine(cr, pv.font, 0, line)

	if pv.readline.Pos >= 0 {
		ch := rune(0)
		if pv.readline.Pos < len(pv.readline.Text) {
			ch = rune(pv.readline.Text[pv.readline.Pos])
		}
		drawCursor(cr, pv.font, 0, pv.readline.Pos+2, ch)
	}
}

func (pv *PromptView) Draw(cr *cairo.Context) {
	pv.marker.Draw(cr)

	cr.Save()
	defer cr.Restore()

	cr.Translate(float64(pv.marker.Width), 0)

	pv.font.Use(cr, false)
	cr.SetSourceRGB(0, 0, 0)
	cr.MoveTo(0, float64(pv.font.ch-pv.font.descent))
	cr.ShowText(string(pv.readline.Text))

	if pv.readline.Pos >= 0 {
		var ext cairo.TextExtents
		cr.TextExtents(string(pv.readline.Text[:pv.readline.Pos]), &ext)
		cr.Rectangle(ext.Width, 0, 3, float64(pv.font.ch-2))
		cr.Fill()
	}
}

func (pv *PromptView) Height() int {
	return pv.font.ch
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
		t := string(pv.readline.Text)
		pv.readline.Text = []byte(t[:pv.cwin.start] + text + t[pv.cwin.end:])
		pv.readline.Pos = pv.cwin.start + len(text)
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
	// log.Printf("usecomp %q %q %q", expand, newText, completions)

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
	x += (len("$ ") + start) * pv.font.cw
	y -= pv.font.ch

	if pv.cwin != nil {
		pv.cwin.Close()
	}
	pv.cwin = NewCompletionWindow(pv, x, y, start, end, completions)
}

func (m *PromptMarker) Draw(cr *cairo.Context) {
	cr.Save()
	defer cr.Restore()

	gray := 0.5
	pad := 2.0
	size := (float64(m.Height) - pad) / 2
	cr.Translate(5, pad)
	cr.SetSourceRGB(gray, gray, gray)
	cr.NewPath()
	cr.MoveTo(0, 0)
	cr.RelLineTo(size, size)
	cr.RelLineTo(-size, size)
	cr.Fill()
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

func (cw *CompletionWindow) accept() {
	cw.pv.OnCompletion(cw.completions[cw.sel])
}

func (cw *CompletionWindow) Key(key keys.Key) bool {
	switch key.Spec() {
	case "Tab":
		if len(cw.completions) == 1 {
			cw.accept()
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
		cw.accept()
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
