package shell

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runBuiltin(t *testing.T, s *Shell, input string) (string, error) {
	cmd, f := s.Run(input)
	assert.Nil(t, cmd)
	out, err := f()
	if err != nil {
		s.Finish(1)
	} else {
		s.Finish(0)
	}
	return out, err
}

func TestAlias(t *testing.T) {
	var err error
	s := NewShell("", map[string]string{}, nil)
	_, err = runBuiltin(t, s, "alias foo xyz")
	assert.Nil(t, err)

	out, _ := runBuiltin(t, s, "alias")
	assert.Contains(t, out, "xyz")
}

func TestCd(t *testing.T) {
	var err error

	cwd, _ := os.Getwd()
	s := NewShell(cwd, map[string]string{}, nil)

	_, err = runBuiltin(t, s, "cd .")
	assert.Nil(t, err)
	assert.Equal(t, s.cwd, cwd)
	assert.Equal(t, s.lastStatus, 0)

	runBuiltin(t, s, "cd testdir")
	assert.Equal(t, s.cwd, cwd+"/testdir")

	runBuiltin(t, s, "cd ..")
	assert.Equal(t, s.cwd, cwd)

	_, err = runBuiltin(t, s, "cd nosuchdir")
	assert.Contains(t, err.Error(), "no such file or dir")
	assert.Equal(t, s.lastStatus, 1)

	runBuiltin(t, s, "cd /")
	assert.Equal(t, s.cwd, "/")

	_, err = runBuiltin(t, s, "cd")
	assert.Contains(t, err.Error(), "no $HOME")

	s.env["HOME"] = "/tmp"
	runBuiltin(t, s, "cd")
	assert.Equal(t, s.cwd, "/tmp")
}

type testCompleter struct {
	expand func(input string) ([]string, error)
}

func (t *testCompleter) Expand(input string) ([]string, error) {
	return t.expand(input)
}

func TestComplete(t *testing.T) {
	c := &testCompleter{}
	s := NewShell("", nil, c)

	c.expand = func(input string) ([]string, error) {
		if input == "p" {
			return []string{"pwd"}, nil
		}
		return nil, fmt.Errorf("unreached")
	}
	comps, err := s.Complete("p")
	assert.Nil(t, err)
	assert.Equal(t, []string{"pwd"}, comps)

	c.expand = func(input string) ([]string, error) {
		if input == "cd f" {
			return []string{"foo", "foox"}, nil
		}
		return nil, fmt.Errorf("unreached")
	}
	comps, err = s.Complete("cd f")
	assert.Nil(t, err)
	assert.Equal(t, []string{"foo", "foox"}, comps)
}
