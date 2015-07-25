package readline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testType(rl *ReadLine, inputs ...string) {
	for _, input := range inputs {
		mod := uint(0)
		if len(input) > 2 && input[:2] == "C-" {
			mod |= Control
			input = input[2:]
		}
		for _, k := range input {
			rl.Key(k, mod)
		}
	}
}

func TestInsert(t *testing.T) {
	rl := NewReadLine()
	testType(rl, "hello, world")
	assert.Equal(t, "hello, world", rl.String())
}

func TestMove(t *testing.T) {
	rl := NewReadLine()
	testType(rl, "hello", "C-a", "x")
	assert.Equal(t, "xhello", rl.String())
}
