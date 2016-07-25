package smash

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/evmar/gocairo/cairo"
	"github.com/kr/pty"

	"github.com/evmar/smash/keys"
	"github.com/evmar/smash/vt100"
)

type TermView struct {
	ViewBase

	termMu sync.Mutex
	term   *vt100.Terminal

	keys io.Writer

	mf *Font

	Running bool
	OnExit  func()
}

func NewTermView(parent View) *TermView {
	return &TermView{
		ViewBase: ViewBase{Parent: parent},
		term:     vt100.NewTerminal(),
		mf:       parent.GetWindow().font,
	}
}

func (t *TermView) Measure(rows, cols int) (width, height int) {
	win := t.GetWindow()
	cr := win.win.GetCairo()
	win.font.Use(cr, false)
	return win.font.cw * 80, win.font.ch * 24
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

func drawText(cr *cairo.Context, mf *Font, x, y int, fg, bg *Color, line string) {
	if bg != nil {
		setColor(cr, bg)
		cr.Rectangle(float64(x), float64(y),
			float64(len(line)*mf.cw), float64(mf.ch))
		cr.Fill()
	}

	cr.MoveTo(float64(x), float64(y+mf.ch-mf.descent))
	setColor(cr, fg)
	cr.ShowText(line)
}

// drawTerminalLine draws one line of a terminal buffer, handling
// layout of text spans of multiple attributes as well as rendering
// the cursor.
func drawTerminalLine(cr *cairo.Context, mf *Font, y int, line []vt100.Cell) {
	var sbuf [100]byte
	buf := sbuf[:]

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
		mf.Use(cr, attr.Bright())

		if bg == &white {
			bg = nil
		}

		drawText(cr, mf, x1*mf.cw, y, fg, bg, string(buf))
	}
}

func drawCursor(cr *cairo.Context, mf *Font, row, col int, ch rune) {
	drawText(cr, mf, col*mf.cw, row*mf.ch, &white, &black, string(ch))
}

func (t *TermView) Draw(cr *cairo.Context) {
	t.mf.Use(cr, false)

	t.termMu.Lock()
	defer t.termMu.Unlock()

	offset := t.term.Top * t.mf.ch
	offset = 0
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

func (t *TermView) Key(key keys.Key) bool {
	if key.Sym == keys.NoSym {
		// Modifier-only keypress.
		return false
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
		return true
	}
	return false
}

func (t *TermView) Scroll(dy int) {
}

func (t *TermView) Height() int {
	t.termMu.Lock()
	defer t.termMu.Unlock()
	lines := len(t.term.Lines)
	if lines > 0 && len(t.term.Lines[lines-1]) == 0 {
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

// unlockingReader wraps Reader by unlocking a mutex when blocked in a
// read.
type unlockingReader struct {
	mu *sync.Mutex
	r  io.Reader
}

func (ur *unlockingReader) Read(buf []byte) (int, error) {
	ur.mu.Unlock()
	n, err := ur.r.Read(buf)
	ur.mu.Lock()
	return n, err
}

func (t *TermView) Start(cmd *exec.Cmd) {
	t.Running = true
	go func() {
		t.runCommand(cmd)
		t.Enqueue(t.Finish)
	}()
}

func (t *TermView) Finish() {
	t.Running = false
	t.term.HideCursor = true
	t.Dirty()
	if t.OnExit != nil {
		t.OnExit()
	}
}

// runCommand executes a subprocess in a TermView and reads its output.
// It should be run in a separate goroutine.
func (t *TermView) runCommand(cmd *exec.Cmd) {
	t.termMu.Lock()
	defer t.termMu.Unlock()
	f, err := pty.Start(cmd)
	if err != nil {
		t.term.DisplayString(err.Error())
		t.Dirty()
		return
	}
	defer f.Close()

	t.term.Height = 24
	t.term.Input = f

	logf, err := os.Create("log")
	if err != nil {
		panic(err)
	}
	defer logf.Close()

	t.keys = f
	logr := &logReader{f, logf}
	lockr := &unlockingReader{r: logr, mu: &t.termMu}
	r := bufio.NewReader(lockr)

	for err == nil {
		err = t.term.Read(r)
		t.Dirty()
	}

	if err != io.EOF {
		// When a pty closes, you get an EIO error instead of an EOF.
		const EIO syscall.Errno = 5
		if perr, ok := err.(*os.PathError); ok {
			if errno, ok := perr.Err.(syscall.Errno); ok && errno == EIO {
				// read /dev/ptmx: input/output error
				err = io.EOF
			}
		}
	}
	if err != io.EOF {
		if err != nil {
			panic(err)
		}
	}
}
