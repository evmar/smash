package main

import (
	"fmt"
	"io"
	"log"
	"sync"
	"unicode/utf8"
)

// xxxx xxIB AAAA CCCC
//  I = inverse
//  B = bold
//  A = background color
//  C = foreground color
type Attr uint16

func (a *Attr) Color() int {
	return int(*a & 0xF)
}
func (a *Attr) SetColor(color int) {
	*a = Attr((uint16(*a) & ^uint16(0xF)) | uint16(color))
}

func (a *Attr) Bold() bool {
	return *a&0x100 != 0
}
func (a *Attr) SetBold(bold bool) {
	flag := uint16(0)
	if bold {
		flag = 1
	}
	*a = Attr((uint16(*a) & ^uint16(0x100)) | (flag << 8))
}

func (a *Attr) Inverse() bool {
	return *a&0x200 != 0
}
func (a *Attr) SetInverse(inverse bool) {
	flag := uint16(0)
	if inverse {
		flag = 1
	}
	*a = Attr((uint16(*a) & ^uint16(0x200)) | (flag << 9))
}

func (a *Attr) BackColor() int {
	return int((*a >> 4) & 0xF)
}
func (a *Attr) SetBackColor(color int) {
	*a = Attr((uint16(*a) & ^uint16(0xF0)) | uint16(color)<<4)
}

func showChar(ch byte) string {
	if ch >= ' ' && ch <= '~' {
		return fmt.Sprintf("'%c'", ch)
	} else {
		return fmt.Sprintf("%#x", ch)
	}
}

type TerminalChar struct {
	Ch   rune
	Attr Attr
}

type Terminal struct {
	Mu     sync.Mutex
	Title  string
	Lines  [][]TerminalChar
	Width  int
	Height int
	Top    int
	Input  io.Writer

	Row, Col int
	Attr     Attr
}

func NewTerminal() *Terminal {
	return &Terminal{
		Lines:  make([][]TerminalChar, 1),
		Row:    0,
		Col:    0,
		Width:  80,
		Height: 24,
	}
}

func promoteEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func (t *Terminal) fixPosition() {
	if t.Row >= t.Top+t.Height {
		t.Top++
	}
	for t.Row >= len(t.Lines) {
		t.Lines = append(t.Lines, make([]TerminalChar, 0))
	}
	for t.Col > len(t.Lines[t.Row]) {
		t.Lines[t.Row] = append(t.Lines[t.Row], TerminalChar{' ', 0})
	}
}

func (t *Terminal) Read(r io.ByteScanner) error {
	c, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch {
	case c == 0x7: // bell
		// ignore
	case c == 0x8: // backspace
		t.Mu.Lock()
		if t.Col > 0 {
			t.Col--
		}
		t.Mu.Unlock()
	case c == 0x1b:
		return t.readEscape(r)
	case c == '\r':
		t.Mu.Lock()
		t.Col = 0
		t.Mu.Unlock()
	case c == '\n':
		t.Mu.Lock()
		t.Col = 0
		t.Row++
		t.fixPosition()
		t.Mu.Unlock()
	case c == '\t':
		t.Mu.Lock()
		t.Col += 8
		t.fixPosition()
		t.Mu.Unlock()
	case c > '~':
		r.UnreadByte()
		return t.readUTF8(r)
	case c >= ' ' && c <= '~':
		t.writeRune(rune(c))
	default:
		log.Printf("term: unhandled %#x", c)
		c = '#'
	}
	return nil
}

func (t *Terminal) writeRune(r rune) {
	t.Mu.Lock()
	if t.Col == t.Width {
		t.Row++
		t.Col = 0
	}
	t.Col++
	t.fixPosition()
	t.Lines[t.Row][t.Col-1] = TerminalChar{r, t.Attr}
	t.Mu.Unlock()
}

func (t *Terminal) readUTF8(r io.ByteScanner) error {
	c, err := r.ReadByte()
	if err != nil {
		return err
	}

	var uc rune
	n := 0
	switch {
	case c&0xE0 == 0xB0:
		uc = rune(c & 0x1F)
		n = 2
	case c&0xF0 == 0xE0:
		uc = rune(c & 0x0F)
		n = 3
	default:
		return fmt.Errorf("term: unhandled utf8 start %#v", c)
	}

	for i := 1; i < n; i++ {
		c, err := r.ReadByte()
		if err != nil {
			return err
		}
		if c&0xC0 != 0x80 {
			return fmt.Errorf("term: unhandled utf8 continuation %#v", c)
		}
		uc = uc<<6 | rune(c&0x3F)
	}
	t.writeRune(uc)
	return nil
}

func (t *Terminal) readEscape(r io.ByteScanner) error {
	// http://invisible-island.net/xterm/ctlseqs/ctlseqs.html
	c, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch {
	case c == '(':
		c, err := r.ReadByte()
		if err != nil {
			return err
		}
		switch c {
		case 'B': // US ASCII
			// ignore
		default:
			log.Printf("TODO: g0 charset %s", showChar(c))
		}
	case c == '=':
		log.Printf("TODO: application keypad")
	case c == '>':
		log.Printf("TODO: normal keypad")
	case c == '[':
		return t.readCSI(r)
	case c == ']':
		// OSC Ps ; Pt BEL
		n, err := t.readInt(r)
		if err != nil {
			return err
		}
		_, err = t.expect(r, ';')
		if err != nil {
			return err
		}
		text, err := t.readTo(r, 0x7)
		if err != nil {
			return err
		}
		switch n {
		case 0, 1, 2:
			t.Mu.Lock()
			t.Title = string(text)
			t.Mu.Unlock()
		default:
			log.Printf("term: bad OSC %d", n)
		}
	case c == 'M': // move up/insert line
		t.Mu.Lock()
		if t.Row == 0 {
			// Insert line above.
			t.Lines = append(t.Lines, nil)
			copy(t.Lines[1:], t.Lines)
			t.Lines[0] = make([]TerminalChar, 0)
		} else {
			if t.Row == t.Top {
				t.Top--
			}
			t.Row--
		}
		t.Mu.Unlock()
	default:
		log.Printf("term: unknown escape %s", showChar(c))
	}
	return nil
}

func readArgs(args []int, values ...*int) {
	for i, val := range values {
		if i < len(args) {
			*val = args[i]
		}
	}
}

func (t *Terminal) readCSI(r io.ByteScanner) error {
	// CSI
	var args []int

	qflag := false
	gtflag := false
L:
	c, err := r.ReadByte()
	if err != nil {
		return err
	}

	switch {
	case c >= '0' && c <= '9':
		r.UnreadByte()
		n, err := t.readInt(r)
		if err != nil {
			return err
		}
		args = append(args, n)

		c, err = r.ReadByte()
		if err != nil {
			return err
		}
		if c == ';' {
			goto L
		}
	case c == '?':
		qflag = true
		goto L
	case c == '>':
		gtflag = true
		goto L
	}

	switch {
	case c == 'C': // cursor forward
		dx := 1
		readArgs(args, &dx)
		t.Mu.Lock()
		t.Col += dx
		t.fixPosition()
		t.Mu.Unlock()
	case c == 'D': // cursor back
		dx := 1
		readArgs(args, &dx)
		t.Mu.Lock()
		t.Col -= dx
		t.fixPosition()
		t.Mu.Unlock()
	case c == 'H': // move to position
		row := 1
		col := 1
		readArgs(args, &row, &col)
		t.Mu.Lock()
		t.Row = t.Top + row - 1
		t.Col = col - 1
		t.fixPosition()
		t.Mu.Unlock()
	case c == 'J': // erase in display
		arg := 0
		readArgs(args, &arg)
		switch arg {
		case 2: // erase all
			t.Mu.Lock()
			t.Lines = t.Lines[:0]
			t.Mu.Unlock()
		default:
			log.Printf("term: unknown erase in display %v", args)
		}
	case c == 'K': // erase in line
		arg := 0
		readArgs(args, &arg)
		switch arg {
		case 0: // erase to right
			t.Mu.Lock()
			t.Lines[t.Row] = t.Lines[t.Row][:t.Col]
			t.Mu.Unlock()
		default:
			log.Printf("term: unknown erase in line %v", args)
		}
	case c == 'P': // erase in line
		arg := 1
		readArgs(args, &arg)
		t.Mu.Lock()
		l := t.Lines[t.Row]
		copy(l[t.Col:], l[t.Col+arg:])
		t.Lines[t.Row] = l[:len(l)-arg]
		t.Mu.Unlock()
	case c == 'd': // line position
		arg := 1
		readArgs(args, &arg)
		t.Mu.Lock()
		t.Row = arg - 1
		t.fixPosition()
		t.Mu.Unlock()
	case (c == 'h' || c == 'l') && !qflag: // reset mode
		reset := c == 'l'
		arg := 0
		readArgs(args, &arg)
		switch arg {
		default:
			log.Printf("TODO: reset mode %d %v", arg, reset)
		}
	case (c == 'h' || c == 'l') && qflag: // DEC private mode set/reset
		reset := c == 'l'
		arg := 0
		readArgs(args, &arg)
		switch arg {
		case 1:
			log.Printf("TODO: application cursor keys mode")
		case 7: // wraparound mode
			log.Printf("TODO: wraparound mode")
		case 12:
			log.Printf("TODO: blinking cursor mode")
		case 25:
			log.Printf("TODO: show cursor mode")
		case 1049: // alternate screen buffer
			log.Printf("TODO: alternate screen buffer")
		default:
			log.Printf("term: unknown dec private mode %v %v", args, reset)
		}
	case c == 'm': // reset
		if len(args) == 0 {
			args = append(args, 0)
		}
		t.Mu.Lock()
		for _, arg := range args {
			switch {
			case arg == 0:
				t.Attr = 0
			case arg == 1:
				t.Attr.SetBold(true)
			case arg == 7:
				t.Attr.SetInverse(true)
			case arg == 27:
				t.Attr.SetInverse(false)
			case arg >= 30 && arg <= 40:
				color := arg - 30
				if color == 9 {
					t.Attr.SetColor(0)
				} else {
					t.Attr.SetColor(color + 1)
				}
			case arg >= 40 && arg <= 50:
				color := arg - 40
				if color == 9 {
					t.Attr.SetBackColor(0)
				} else {
					t.Attr.SetBackColor(color + 1)
				}
			default:
				log.Printf("term: unknown color %v", args)
			}
		}
		t.Mu.Unlock()
	case gtflag && c == 'n': // disable modifiers
		arg := 2
		readArgs(args, &arg)
		switch arg {
		case 0:
			log.Printf("TODO: disable modify keyboard")
		case 1:
			log.Printf("TODO: disable modify cursor keys")
		case 2:
			log.Printf("TODO: disable modify function keys")
		case 4:
			log.Printf("TODO: disable modify other keys")
		}
	case c == 'n': // device status report
		arg := 0
		readArgs(args, &arg)
		switch arg {
		case 5:
			_, err := t.Input.Write([]byte("\x1b[0n"))
			return err
		case 6:
			t.Mu.Lock()
			pos := fmt.Sprintf("\x1b[%d;%dn", t.Row+1, t.Col+1)
			t.Mu.Unlock()
			_, err := t.Input.Write([]byte(pos))
			return err
		default:
			log.Printf("term: unknown status report arg %v", args)
		}
	case c == 'r': // set scrolling region
		top, bot := 1, 1
		readArgs(args, &top, &bot)
		if top == 1 && bot == t.Height {
			// Just setting the current region as scroll.
		} else {
			log.Printf("TODO: set scrolling region %v", args)
		}
	default:
		log.Printf("term: unknown CSI %v %s", args, showChar(c))
	}
	return nil
}

func (t *Terminal) expect(r io.ByteScanner, exp byte) (bool, error) {
	c, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	ok := c == exp
	if !ok {
		log.Printf("expect %s failed, got %s", showChar(exp), showChar(c))
	}
	return ok, nil
}

func (t *Terminal) readInt(r io.ByteScanner) (int, error) {
	n := 0
	for i := 0; i < 20; i++ {
		c, err := r.ReadByte()
		if err != nil {
			return -1, err
		}
		if c >= '0' && c <= '9' {
			n = n*10 + int(c) - '0'
		} else {
			r.UnreadByte()
			return n, err
		}
	}
	return -1, fmt.Errorf("term: readInt overlong")
}

func (t *Terminal) readTo(r io.ByteScanner, end byte) ([]byte, error) {
	var buf []byte
	for i := 0; i < 1000; i++ {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c == end {
			return buf, nil
		}
		buf = append(buf, c)
	}
	return nil, fmt.Errorf("term: readTo(%s) overlong", showChar(end))
}

func (t *Terminal) ToString() string {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	var buf [6]byte
	str := ""
	for _, l := range t.Lines {
		if str != "" {
			str += "\n"
		}
		for _, c := range l {
			n := utf8.EncodeRune(buf[:], c.Ch)
			str += string(buf[:n])
		}
	}
	return str
}
