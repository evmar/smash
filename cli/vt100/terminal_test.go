package vt100

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeInput(input string) *bufio.Reader {
	buf := strings.NewReader(input)
	return bufio.NewReader(buf)
}

func newTestTerminal() (*Terminal, *TermReader) {
	term := NewTerminal()
	tr := NewTermReader(func(f func(t *Terminal)) {
		f(term)
	})
	return term, tr
}

func mustRun(t *testing.T, tr *TermReader, input string) {
	r := makeInput(input)
	var err error
	for err == nil {
		err = tr.Read(r)
	}
	assert.Equal(t, err, io.EOF)
}

func assertPos(t *testing.T, term *Terminal, row, col int) {
	assert.Equal(t, row, term.Row)
	assert.Equal(t, col, term.Col)
}

func TestBasic(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "test")
	assert.Equal(t, "test", term.ToString())
	mustRun(t, tr, "\nbar")
	assert.Equal(t, "test\nbar", term.ToString())
	mustRun(t, tr, "\rfoo")
	assert.Equal(t, "test\nfoo", term.ToString())
	mustRun(t, tr, "\n\n")
	assert.Equal(t, "test\nfoo\n\n", term.ToString())
	mustRun(t, tr, "x\ty")
	assert.Equal(t, "test\nfoo\n\nx       y", term.ToString())
}

func TestTitle(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x1b]0;title\x07text")
	assert.Equal(t, "title", term.Title)
	assert.Equal(t, "text", term.ToString())
}

func TestReset(t *testing.T) {
	term, tr := newTestTerminal()
	tr.Attr = 43
	mustRun(t, tr, "\x1b[0m")
	assert.Equal(t, Attr(0), tr.Attr)
	assert.Equal(t, "", term.ToString())
}

func TestColor(t *testing.T) {
	term, tr := newTestTerminal()
	assert.Equal(t, false, tr.Attr.Bright())
	assert.Equal(t, false, tr.Attr.Inverse())
	assert.Equal(t, 0, tr.Attr.Color())

	mustRun(t, tr, "\x1b[1;34m")
	assert.Equal(t, true, tr.Attr.Bright())
	assert.Equal(t, 5, tr.Attr.Color())
	assert.Equal(t, "", term.ToString())

	mustRun(t, tr, "\x1b[7m")
	assert.Equal(t, true, tr.Attr.Inverse())

	mustRun(t, tr, "\x1b[m")
	assert.Equal(t, Attr(0), tr.Attr)
}

func TestBackspace(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x08")
	assert.Equal(t, "", term.ToString())
	mustRun(t, tr, "x\x08")
	assert.Equal(t, "x", term.ToString())
	mustRun(t, tr, "ab\x08c")
	assert.Equal(t, "ac", term.ToString())
}

func TestEraseLine(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "hello")
	term.Col -= 2
	mustRun(t, tr, "\x1b[K")
	assert.Equal(t, "hel", term.ToString())
	mustRun(t, tr, "\x1b[1K")
	assert.Equal(t, "   ", term.ToString())
}

func TestEraseDisplay(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "hellofoo\b\b\b")
	mustRun(t, tr, "\x1b[J")
	assert.Equal(t, "hello", term.ToString())
	mustRun(t, tr, "\x1b[2J")
	assert.Equal(t, "", term.ToString())
}

func TestDelete(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "abcdef\x08\x08\x08\x1b[1P")
	assert.Equal(t, "abcef", term.ToString())

	// Check deleting past the end of the line.
	mustRun(t, tr, "\x1b[5P")
	assert.Equal(t, "abc", term.ToString())
}

func TestBell(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x07")
	// ignored
	assert.Equal(t, "", term.ToString())
}

func TestPrivateModes(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x1b[?1049h")
	// ignored
	assert.Equal(t, "", term.ToString())

	mustRun(t, tr, "\x1b[?7h")
	// ignored
	assert.Equal(t, "", term.ToString())
}

func TestScrollingRegion(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x1b[1;24r")
	// ignored
	assert.Equal(t, "", term.ToString())
}

func TestResetMode(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x1b[4l")
	// ignored
	assert.Equal(t, "", term.ToString())
}

func TestMoveTo(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "hello\x1b[HX")
	assert.Equal(t, "Xello", term.ToString())
	mustRun(t, tr, "\x1b[1;3HX")
	assert.Equal(t, "XeXlo", term.ToString())
	mustRun(t, tr, "\x1b[0;0HY")
	assert.Equal(t, "YeXlo", term.ToString())
}

func TestMoveToLine(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "hello\n\n\x1b[2dfoo")
	assert.Equal(t, "hello\nfoo\n", term.ToString())
}

func TestCursor(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "foo\nbar")
	assertPos(t, term, 1, 3)

	mustRun(t, tr, "\x1b[C")
	assertPos(t, term, 1, 4)
	mustRun(t, tr, "\x1b[2C")
	assertPos(t, term, 1, 6)
	assert.Equal(t, "foo\nbar   ", term.ToString())

	mustRun(t, tr, "\x1b[A!")
	assertPos(t, term, 0, 7)
	assert.Equal(t, "foo   !\nbar   ", term.ToString())

	mustRun(t, tr, "\x1b[5D")
	assertPos(t, term, 0, 2)
}

func TestScrollUp(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "aaa\nbbb\nccc\n")
	assertPos(t, term, 3, 0)
	mustRun(t, tr, "\x1bMX")
	assert.Equal(t, "aaa\nbbb\nXcc\n", term.ToString())
	mustRun(t, tr, "\x1bMY")
	mustRun(t, tr, "\x1bMZ")
	assert.Equal(t, "aaZ\nbYb\nXcc\n", term.ToString())
	mustRun(t, tr, "\x1bM1")
	assert.Equal(t, "   1\naaZ\nbYb\nXcc\n", term.ToString())
}

func TestScrollUpDropLines(t *testing.T) {
	term, tr := newTestTerminal()
	term.Height = 3
	mustRun(t, tr, "aaa\nbbb\nccc\n")
	assert.Equal(t, "aaa\nbbb\nccc\n", term.ToString())
	mustRun(t, tr, "\x1bM\x1bM\x1bM\x1bMx")
	assert.Equal(t, "x\naaa\nbbb", term.ToString())
}

func TestWrap(t *testing.T) {
	term, tr := newTestTerminal()
	term.Width = 5
	mustRun(t, tr, "1234567890")
	assert.Equal(t, "12345\n67890", term.ToString())
}

func TestUTF8(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\xe2\x96\xbd")
	assert.Equal(t, rune(0x25bd), term.Lines[0][0].Ch)
}

func TestStatusReport(t *testing.T) {
	term, tr := newTestTerminal()
	buf := &bytes.Buffer{}
	tr.Input = buf
	mustRun(t, tr, "\x1b[5n")
	assert.Equal(t, "", term.ToString())
	assert.Equal(t, "\x1b[0n", buf.String())

	buf.Reset()
	mustRun(t, tr, "\x1b[6n")
	assert.Equal(t, "\x1b[1;1R", buf.String())
}

func TestCSIDisableModifiers(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x1b[>0n")
	assert.Equal(t, "", term.ToString())
	// TODO: implement the disabling, whatever that is.
}

func TestSendDeviceAttributes(t *testing.T) {
	term, tr := newTestTerminal()
	buf := &bytes.Buffer{}
	tr.Input = buf
	mustRun(t, tr, "\x1b[c")
	assert.Equal(t, "", term.ToString())
	assert.Equal(t, "", buf.String())
	mustRun(t, tr, "\x1b[>c")
	assert.Equal(t, "", term.ToString())
	assert.Equal(t, "\x1b[0;0;0c", buf.String())
}

func TestHideCursor(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "\x1b[?25l")
	assert.Equal(t, true, term.HideCursor)
	mustRun(t, tr, "\x1b[?25h")
	assert.Equal(t, false, term.HideCursor)
}

func TestInsertBlanks(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "ABC\b\b\x1b[@x")
	assert.Equal(t, "AxBC", term.ToString())
	mustRun(t, tr, "\x1b[2@y")
	assert.Equal(t, "Axy BC", term.ToString())
}

func TestInsertLine(t *testing.T) {
	term, tr := newTestTerminal()
	mustRun(t, tr, "foo\nbar\nbaz\n")
	mustRun(t, tr, "\x1b[2A\x1b[L") // two lines up, insert line
	mustRun(t, tr, "\nX")
	assert.Equal(t, "foo\n\nXar\nbaz\n", term.ToString())
}

func TestBinary(t *testing.T) {
	term, tr := newTestTerminal()
	// Don't choke on non-UTF8 inputs.
	// TODO: maybe render them with some special character to represent
	// mojibake.
	mustRun(t, tr, "\xc8\x00\x64\x00")
	assert.Equal(t, "@@d@", term.ToString())
}

func TestAllColors(t *testing.T) {
	buf := &bytes.Buffer{}
	for i := 30; i < 50; i++ {
		fmt.Fprintf(buf, "\x1b[%dmx", i)
	}
	term, tr := newTestTerminal()
	mustRun(t, tr, buf.String())
	x20 := "xxxxxxxxxx" + "xxxxxxxxxx"
	assert.Equal(t, x20, term.ToString())
	assert.Nil(t, term.Validate())
}
