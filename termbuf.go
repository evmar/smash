package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"smash/base"
	"smash/keys"
	"syscall"
	"time"

	"github.com/kr/pty"
	"github.com/martine/gocairo/cairo"
)

type Metrics struct {
	// Character width and height.
	cw, ch int
	// Adjustment from drawing baseline to bottom of character.
	descent int
}

type TermBuf struct {
	ViewBase
	term *Terminal

	keys io.Writer

	metrics Metrics

	offset     int
	scrollRows int
	anim       *base.Lerp
}

func NewTermBuf(parent View) *TermBuf {
	return &TermBuf{
		ViewBase: ViewBase{Parent: parent},
		term:     NewTerminal(),
	}
}

func ansiColor(c int, bold bool, def *Color) *Color {
	if c == 0 {
		return def
	}
	if bold {
		return &ansiBrightColors[c-1]
	}
	return &ansiColors[c-1]
}

func setColor(cr *cairo.Context, color *Color) {
	cr.SetSourceRGB(float64(color.R)/0xff, float64(color.G)/0xff, float64(color.B)/0xff)
}

// drawTerminalLine draws one line of a terminal buffer, handling
// layout of text spans of multiple attributes as well as rendering
// the cursor.
func drawTerminalLine(cr *cairo.Context, metrics *Metrics, y int, line []TerminalChar, cursorCol int) bool {
	drewCursor := false

	cw := metrics.cw
	ch := metrics.ch
	descent := metrics.descent

	// TODO: reuse buf across lines?
	buf := make([]byte, 0, 100)

	// Collect spans of text with the same attributes, to batch
	// the drawing calls to cairo.  The cursor is specially handled as
	// a separate span.
	var x1, x2 int
	for x1 = 0; x1 < len(line); x1 = x2 {
		buf = buf[:0]
		attr := line[x1].Attr

		inCursor := false
		for x2 = x1; x2 < len(line) && line[x2].Attr == attr && !inCursor; x2++ {
			if x1 == cursorCol {
				// This span is the cursor.
				inCursor = true
			} else if x2 == cursorCol {
				// Hit the cursor; let the next span handle it.
				break
			}
			ch := line[x2].Ch
			if ch < 0x7f {
				buf = append(buf, byte(ch))
			} else {
				log.Printf("TODO: render unicode")
				buf = append(buf, '#')
			}
		}

		fg := ansiColor(attr.Color(), attr.Bold(), &black)
		bg := ansiColor(attr.BackColor(), false, &white)
		if attr.Inverse() {
			fg, bg = bg, fg
		}

		if inCursor {
			fg = &white
			bg = &black
			drewCursor = true
		}

		if bg != &white {
			setColor(cr, bg)
			cr.Rectangle(float64(x1*cw), float64(y),
				float64(len(buf)*cw), float64(ch))
			cr.Fill()
		}

		cr.MoveTo(float64(x1*cw), float64(y+ch-descent+1))
		setColor(cr, fg)
		cr.ShowText(string(buf))
	}

	return drewCursor
}

func (t *TermBuf) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()

	cr.SetSourceRGB(0, 0, 0)
	cr.SelectFontFace("monospace", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(14)
	if t.metrics.cw == 0 {
		ext := cairo.FontExtents{}
		cr.FontExtents(&ext)
		t.metrics.cw = int(ext.MaxXAdvance)
		t.metrics.ch = int(ext.Height)
		t.metrics.descent = int(ext.Descent)
	}

	t.term.Mu.Lock()
	defer t.term.Mu.Unlock()

	offset := (t.term.Top + t.scrollRows) * t.metrics.ch

	cr.IdentityMatrix()
	cr.Translate(0, float64(-t.offset))
	if t.offset != offset {
		if t.anim != nil && t.anim.Done {
			t.anim = nil
		}
		if t.anim == nil {
			t.anim = base.NewLerp(&t.offset, offset, 40*time.Millisecond)
			anims.Add(t.anim)
		} else {
			// TODO adjust existing anim
		}
	}
	firstLine := t.offset / t.metrics.ch
	if firstLine < 0 {
		firstLine = 0
	}
	lastLine := t.term.Top + t.term.Height
	if lastLine > len(t.term.Lines) {
		lastLine = len(t.term.Lines)
	}

	drewCursor := false
	for row := firstLine; row < lastLine; row++ {
		cursorCol := -1
		if row == t.term.Row {
			cursorCol = t.term.Col
		}
		if drawTerminalLine(cr, &t.metrics, row*t.metrics.ch, t.term.Lines[row], cursorCol) {
			drewCursor = true
		}
	}

	if !t.term.HideCursor && !drewCursor {
		setColor(cr, &black)
		cr.Rectangle(float64(t.term.Col*t.metrics.cw),
			float64(t.term.Row*t.metrics.ch),
			float64(t.metrics.cw), float64(t.metrics.ch))
		cr.Fill()
	}
}

func (t *TermBuf) Key(key keys.Key) {
	if key.Sym == keys.NoSym {
		// Modifier-only keypress.
		return
	}

	// log.Printf("key %#x %c", key, key)
	var send string
	if key.Sym < keys.FirstNonASCIISym {
		ch := byte(key.Sym)
		if key.Mods&keys.ModControl != 0 {
			// Ctl: C-a means "send keycode 1".
			ch = ch - 'a' + 1
		}
		if key.Mods&keys.ModMeta != 0 {
			// Alt: send an escape before the next key.
			send += "\x1b"
		}
		send += fmt.Sprintf("%c", ch)
	} else {
		switch key.Sym {
		case keys.Up:
			send = "\x1b[A"
		case keys.Down:
			send = "\x1b[B"
		case keys.Right:
			send = "\x1b[C"
		case keys.Left:
			send = "\x1b[D"
		default:
			log.Printf("unhandled key %#v", key)
		}
	}

	if send != "" {
		io.WriteString(t.keys, send)
		t.Dirty()
	}
}

func (t *TermBuf) Scroll(dy int) {
	t.scrollRows -= dy
	t.Dirty()
}

type logReader struct {
	io.Reader
	log io.Writer
}

func (lr *logReader) Read(buf []byte) (int, error) {
	n, err := lr.Reader.Read(buf)
	lr.log.Write(buf[:n])
	return n, err
}

// runBash runs bash in the TermBuf, blocking until it completes.
func (t *TermBuf) runBash() {
	logf, err := os.Create("log")
	check(err)

	cmd := exec.Command("bash")
	f, err := pty.Start(cmd)
	check(err)

	t.term.Height = 24
	t.term.Input = f

	t.keys = f
	lr := &logReader{f, logf}
	r := bufio.NewReader(lr)

	for err == nil {
		err = t.term.Read(r)
		t.Dirty()
	}
	logf.Close()

	if err != io.EOF {
		if perr, ok := err.(*os.PathError); ok {
			if errno, ok := perr.Err.(syscall.Errno); ok && errno == EIO {
				// read /dev/ptmx: input/output error
				err = io.EOF
			}
		}
	}
	if err != io.EOF {
		check(err)
	}
}
