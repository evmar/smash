package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCd(t *testing.T) {
	cwd, _ := os.Getwd()
	s := NewShell(cwd, map[string]string{})

	s.builtinCd([]string{"cd", "."})
	assert.Equal(t, s.cwd, cwd)

	s.builtinCd([]string{"cd", "basic"})
	assert.Equal(t, s.cwd, cwd+"/basic")

	s.builtinCd([]string{"cd", ".."})
	assert.Equal(t, s.cwd, cwd)

	s.builtinCd([]string{"cd", "nosuchdir"})
	assert.Equal(t, s.cwd, cwd)

	s.builtinCd([]string{"cd", "/tmp"})
	assert.Equal(t, s.cwd, "/tmp")
}
