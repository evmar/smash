package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"smash/base"
	"syscall"
	"time"

	"github.com/kr/pty"
	"github.com/martine/gocairo/cairo"
)

type TermBuf struct {
	win  *Window
	term *Terminal

	keys io.Writer

	// Character metrics, in pixels.
	cw, ch  int
	descent int

	offset     int
	scrollRows int
	anim       *base.Lerp
}

func NewTermBuf(win *Window) *TermBuf {
	return &TermBuf{
		win:  win,
		term: NewTerminal(),
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

func (t *TermBuf) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()

	cr.SetSourceRGB(0, 0, 0)
	cr.SelectFontFace("monospace", cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(14)
	if t.cw == 0 {
		ext := &cairo.FontExtents{}
		cr.FontExtents(ext)
		log.Printf("font extents %#v", ext)
		t.cw = int(ext.MaxXAdvance)
		t.ch = int(ext.Height)
		t.descent = int(ext.Descent)
	}

	t.term.Mu.Lock()
	defer t.term.Mu.Unlock()

	offset := (t.term.Top + t.scrollRows) * t.ch

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
	firstLine := t.offset / t.ch
	if firstLine < 0 {
		firstLine = 0
	}
	lastLine := t.term.Top + t.term.Height
	if lastLine > len(t.term.Lines) {
		lastLine = len(t.term.Lines)
	}

	buf := make([]byte, 0, 80)
	drewCursor := false
	for row := firstLine; row < lastLine; row++ {
		line := t.term.Lines[row]

		// Collect spans of text with the same attributes, to batch
		// the drawing calls to cairo.  The cursor is specially handled as
		// a separate batch.
		var x1, x2 int
		for x1 = 0; x1 < len(line); x1 = x2 {
			buf = buf[:0]
			attr := line[x1].Attr

			inCursor := false
			for x2 = x1; x2 < len(line) && line[x2].Attr == attr && !inCursor; x2++ {
				if !t.term.HideCursor &&
					row == t.term.Row {
					if x1 == t.term.Col {
						// This span is the cursor.
						inCursor = true
					} else if x2 == t.term.Col {
						// Hit the cursor; let the next span handle it.
						break
					}
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
				cr.Rectangle(float64(x1*t.cw), float64(row*t.ch),
					float64(len(buf)*t.cw), float64(t.ch))
				cr.Fill()
			}

			cr.MoveTo(float64(x1*t.cw), float64((row+1)*t.ch-t.descent+1))
			setColor(cr, fg)
			cr.ShowText(string(buf))
		}
	}

	if !t.term.HideCursor && !drewCursor {
		setColor(cr, &black)
		cr.Rectangle(float64(t.term.Col*t.cw),
			float64(t.term.Row*t.ch),
			float64(t.cw), float64(t.ch))
		cr.Fill()
	}
}

func (t *TermBuf) Key(key *base.Key) {
	// log.Printf("key %#x %c", key, key)
	var send string
	if key.Text != "" {
		send = key.Text
	} else if key.Special != base.KeyNone {
		switch key.Special {
		case base.KeyUp:
			send = "\x1b[A"
		case base.KeyDown:
			send = "\x1b[B"
		case base.KeyRight:
			send = "\x1b[C"
		case base.KeyLeft:
			send = "\x1b[D"
		default:
			log.Printf("key %#v", key)
		}
	}

	if send != "" {
		io.WriteString(t.keys, send)
		t.win.xwin.Dirty()
	}
}

func (t *TermBuf) Scroll(dy int) {
	t.scrollRows -= dy
	t.win.xwin.Dirty()
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
		t.win.xwin.Dirty()
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
