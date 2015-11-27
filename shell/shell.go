package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Shell struct {
	cwd        string
	env        map[string]string
	lastStatus int
}

func NewShell(cwd string, env map[string]string) *Shell {
	return &Shell{
		cwd: cwd,
		env: env,
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

func (s *Shell) Run(argv []string) (*exec.Cmd, error) {
	if len(argv) == 0 {
		return nil, nil
	}

	var cmd *exec.Cmd
	var err error
	switch argv[0] {
	case "cd":
		err = s.builtinCd(argv)
	default:
		cmd = exec.Command(argv[0], argv[1:]...)
		cmd.Dir = s.cwd
	}
	if err != nil {
		s.Finish(1)
	}
	return cmd, err
}

func (s *Shell) Finish(status int) {
	s.lastStatus = status
}
