package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Delegate interface {
	OnShellError(string)
	OnShellStart(cwd string, argv []string) error
}

type Shell struct {
	delegate   Delegate
	cwd        string
	env        map[string]string
	lastStatus int
}

func NewShell(delegate Delegate, cwd string, env map[string]string) *Shell {
	return &Shell{
		delegate: delegate,
		cwd:      cwd,
		env:      env,
	}
}

func (s *Shell) Parse(cmd string) []string {
	return strings.Split(cmd, " ")
}

func (s *Shell) builtinCd(argv []string) error {

	var dir string
	switch len(argv) {
	case 1:
		var ok bool
		dir, ok = s.env["HOME"]
		if !ok {
			return fmt.Errorf("cd: no $HOME")
		}
	case 2:
		dir = argv[1]
	default:
		return fmt.Errorf("usage: cd [dir]")
	}

	dir = filepath.Join(s.cwd, dir)
	st, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("%q: not a directory", dir)
	}
	s.cwd = dir
	return nil
}

func (s *Shell) Run(argv []string) {
	if len(argv) == 0 {
		return
	}

	var err error
	switch argv[0] {
	case "cd":
		err = s.builtinCd(argv)
	default:
		err = s.delegate.OnShellStart(s.cwd, argv)
	}
	if err != nil {
		s.delegate.OnShellError(err.Error())
		s.Finish(1)
	}
}

func (s *Shell) Finish(status int) {
	s.lastStatus = status
}
