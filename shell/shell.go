package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Completer interface {
	Complete(input string) (int, []string, error)
}

type Shell struct {
	cwd        string
	env        map[string]string
	aliases    map[string]string
	lastStatus int

	completer Completer
}

type Builtin func() (string, error)

func NewShell(cwd string, env map[string]string, completer Completer) *Shell {
	return &Shell{
		cwd: cwd,
		env: env,
		aliases: map[string]string{
			"ls":   "ls --color",
			"grep": "grep --color",
		},
		completer: completer,
	}
}

func (s *Shell) Complete(input string) (int, []string, error) {
	// if err := s.bash.Chdir(s.cwd); err != nil {
	// 	return nil, err
	// }
	return s.completer.Complete(input)
}

func (s *Shell) builtinAlias(argv []string) (string, error) {
	buf := &bytes.Buffer{}
	switch len(argv) {
	case 1:
		cmds := []string{}
		maxLen := 0
		for cmd := range s.aliases {
			cmds = append(cmds, cmd)
			if len(cmd) > maxLen {
				maxLen = len(cmd)
			}
		}
		sort.Strings(cmds)
		for _, cmd := range cmds {
			fmt.Fprintf(buf, "%*s %s\n", maxLen, cmd, s.aliases[cmd])
		}
	case 3:
		s.aliases[argv[1]] = argv[2]
	default:
		return "", fmt.Errorf("usage: alias [input output]")
	}
	return buf.String(), nil
}

func (s *Shell) builtinCd(argv []string) (string, error) {
	var dir string
	switch len(argv) {
	case 1:
		var ok bool
		dir, ok = s.env["HOME"]
		if !ok {
			return "", fmt.Errorf("cd: no $HOME")
		}
	case 2:
		dir = argv[1]
	default:
		return "", fmt.Errorf("usage: cd [dir]")
	}

	if dir[0] == '/' {
		// absolute path
	} else {
		dir = filepath.Join(s.cwd, dir)
	}
	st, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%q: not a directory", dir)
	}
	s.cwd = dir
	return "", nil
}

func (s *Shell) parse(input string) []string {
	argv := strings.Split(input, " ")
	if len(argv) == 1 && argv[0] == "" {
		return nil
	}

	if alias := s.aliases[argv[0]]; alias != "" {
		exp := strings.Split(alias, " ")
		argv = append(exp, argv[1:]...)
	}
	return argv
}

func (s *Shell) Run(input string) (*exec.Cmd, Builtin) {
	argv := s.parse(input)
	if argv == nil {
		return nil, nil
	}

	var cmd *exec.Cmd
	var builtin Builtin
	switch argv[0] {
	case "alias":
		builtin = func() (string, error) { return s.builtinAlias(argv) }
	case "cd":
		builtin = func() (string, error) { return s.builtinCd(argv) }
	default:
		cmd = exec.Command(argv[0], argv[1:]...)
		cmd.Dir = s.cwd
	}
	return cmd, builtin
}

func (s *Shell) Finish(status int) {
	s.lastStatus = status
}
