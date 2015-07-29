package readline

import (
	"smash/keys"
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
