package readline

import (
	"github.com/evmar/smash/keys"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testType(rl *ReadLine, inputs ...string) {
	for _, input := range inputs {
		mod := uint(0)
		if len(input) > 2 && input[:2] == "C-" {
			mod |= keys.ModControl
			input = input[2:]
		}
		for _, k := range input {
			rl.Key(keys.Key{keys.Sym(k), mod})
		}
	}
}

func TestInsert(t *testing.T) {
	rl := NewConfig().NewReadLine()
	testType(rl, "hello, world")
	assert.Equal(t, "hello, world", rl.String())
}

func TestMove(t *testing.T) {
	rl := NewConfig().NewReadLine()
	testType(rl, "hello", "C-a", "X")
	assert.Equal(t, "Xhello", rl.String())
	testType(rl, "C-f", "C-f", "Y")
	assert.Equal(t, "XheYllo", rl.String())
}

func TestClear(t *testing.T) {
	rl := NewConfig().NewReadLine()
	testType(rl, "hello", "C-k")
	assert.Equal(t, "hello", rl.String())
	testType(rl, "C-b", "C-b", "C-k")
	assert.Equal(t, "hel", rl.String())
	testType(rl, "C-b", "C-u")
	assert.Equal(t, "l", rl.String())
}

func TestBackspace(t *testing.T) {
	rl := NewConfig().NewReadLine()
	testType(rl, "hello", "C-h")
	assert.Equal(t, "hell", rl.String())
	testType(rl, "C-b", "C-b", "C-h", "C-h")
	assert.Equal(t, "ll", rl.String())
}

func TestAccept(t *testing.T) {
	rl := NewConfig().NewReadLine()
	ok := false
	rl.Accept = func(text string) bool {
		ok = true
		return true
	}
	testType(rl, "hello")
	assert.Equal(t, false, ok)
	testType(rl, "\r")
	assert.Equal(t, true, ok)
}

func TestHistory(t *testing.T) {
	rl := NewConfig().NewReadLine()

	// Don't die when there's no history.
	testType(rl, "C-n", "C-p")

	rl.Accept = func(text string) bool {
		rl.Clear()
		return true
	}
	testType(rl, "foo\r", "bar\r")
	testType(rl, "C-p")
	assert.Equal(t, "bar", rl.String())
	testType(rl, "C-p")
	assert.Equal(t, "foo", rl.String())
	testType(rl, "C-p")
	assert.Equal(t, "foo", rl.String()) // didn't loop around
	testType(rl, "C-n")
	assert.Equal(t, "bar", rl.String())
	testType(rl, "C-n")
	assert.Equal(t, "bar", rl.String()) // didn't advance
}
