package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"smash/keys"
	"syscall"

	"github.com/kr/pty"
	"github.com/martine/gocairo/cairo"
)

type TermBuf struct {
	ViewBase
	term *Terminal

	keys io.Writer

	mf *MonoFont

	OnExit func()
}

func NewTermBuf(parent View) *TermBuf {
	return &TermBuf{
		ViewBase: ViewBase{Parent: parent},
		term:     NewTerminal(),
		mf:       GetMonoFont(),
	}
}

func ansiColor(c int, bright bool, def *Color) *Color {
	if c == 0 {
		return def
	}
	if bright {
		return &ansiBrightColors[c-1]
	}
	return &ansiColors[c-1]
}

func setColor(cr *cairo.Context, color *Color) {
	cr.SetSourceRGB(float64(color.R)/0xff, float64(color.G)/0xff, float64(color.B)/0xff)
}

func drawText(cr *cairo.Context, mf *MonoFont, x, y int, fg, bg *Color, line string) {
	if bg != nil {
		setColor(cr, bg)
		cr.Rectangle(float64(x), float64(y),
			float64(len(line)*mf.cw), float64(mf.ch))
		cr.Fill()
	}

	cr.MoveTo(float64(x), float64(y+mf.ch-mf.descent+1))
	setColor(cr, fg)
	cr.ShowText(line)
}

// drawTerminalLine draws one line of a terminal buffer, handling
// layout of text spans of multiple attributes as well as rendering
// the cursor.
func drawTerminalLine(cr *cairo.Context, mf *MonoFont, y int, line []TerminalChar) {
	// TODO: reuse buf across lines?
	buf := make([]byte, 0, 100)

	// Collect spans of text with the same attributes, to batch
	// the drawing calls to cairo.
	var x1, x2 int
	for x1 = 0; x1 < len(line); x1 = x2 {
		buf = buf[:0]
		attr := line[x1].Attr

		for x2 = x1; x2 < len(line) && line[x2].Attr == attr; x2++ {
			ch := line[x2].Ch
			if ch < 0x7f {
				buf = append(buf, byte(ch))
			} else {
				log.Printf("TODO: render unicode")
				buf = append(buf, '#')
			}
		}

		fg := ansiColor(attr.Color(), attr.Bright(), &black)
		bg := ansiColor(attr.BackColor(), false, &white)
		if attr.Inverse() {
			fg, bg = bg, fg
		}

		if bg == &white {
			bg = nil
		}

		drawText(cr, mf, x1*mf.cw, y, fg, bg, string(buf))
	}
}

func drawCursor(cr *cairo.Context, mf *MonoFont, row, col int, ch rune) {
	drawText(cr, mf, col*mf.cw, row*mf.ch, &white, &black, string(ch))
}

func (t *TermBuf) Draw(cr *cairo.Context) {
	t.mf.Use(cr)

	t.term.Mu.Lock()
	defer t.term.Mu.Unlock()

	offset := t.term.Top * t.mf.ch
	if offset > 0 {
		cr.Save()
		defer cr.Restore()
		cr.Translate(0, float64(-offset))
	}

	firstLine := offset / t.mf.ch
	if firstLine < 0 {
		firstLine = 0
	}
	lastLine := t.term.Top + t.term.Height
	if lastLine > len(t.term.Lines) {
		lastLine = len(t.term.Lines)
	}

	for row := firstLine; row < lastLine; row++ {
		drawTerminalLine(cr, t.mf, row*t.mf.ch, t.term.Lines[row])
	}

	if !t.term.HideCursor {
		ch := rune(0)
		if t.term.Row < len(t.term.Lines) &&
			t.term.Col < len(t.term.Lines[t.term.Row]) {
			ch = t.term.Lines[t.term.Row][t.term.Col].Ch
		}
		drawCursor(cr, t.mf, t.term.Row, t.term.Col, ch)
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
}

func (t *TermBuf) Height() int {
	t.term.Mu.Lock()
	defer t.term.Mu.Unlock()
	lines := len(t.term.Lines) - t.term.Top
	if len(t.term.Lines) > 0 && len(t.term.Lines[len(t.term.Lines)-1]) == 0 {
		// Drop the trailing newline.
		lines--
	}
	return lines * t.mf.ch
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

func (t *TermBuf) Start(cmd *exec.Cmd) {
	go func() {
		t.runCommand(cmd)
		t.Enqueue(t.Exit)
	}()
}

func (t *TermBuf) Exit() {
	t.term.HideCursor = true
	t.Dirty()
	if t.OnExit != nil {
		t.OnExit()
	}
}

func (t *TermBuf) runCommand(cmd *exec.Cmd) {
	logf, err := os.Create("log")
	check(err)

	f, err := pty.Start(cmd)
	check(err)
	defer f.Close()

	t.term.Mu.Lock()
	t.term.Height = 24
	t.term.Input = f
	t.term.Mu.Unlock()

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
