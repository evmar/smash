// Package bash wraps a bash subprocess, reusing its tab completion support.
package bash

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/kr/pty"
)

type winsize struct {
	row, col   uint16
	xpix, ypix uint16
}

type Bash struct {
	cmd   *exec.Cmd
	pty   *os.File
	lines *bufio.Scanner
}

// Magic string indicating a ready prompt.
const promptMagic = "***READY>"

// inputrc file contents, used to make completion always display and
// disable the pager.
const inputrc = `set completion-query-items 0
set page-completions off
`

// StartBash starts up a new bash subprocess for use in completions.
func StartBash() (b *Bash, err error) {
	f, err := ioutil.TempFile("", "smash-inputrc")
	if err != nil {
		return nil, err
	}
	io.WriteString(f, inputrc)
	f.Close()
	defer os.Remove(f.Name())

	b = &Bash{}
	b.cmd = exec.Command("bash", "-i")
	b.cmd.Env = append(b.cmd.Env, fmt.Sprintf("INPUTRC=%s", f.Name()))
	if b.pty, err = pty.Start(b.cmd); err != nil {
		return nil, err
	}
	b.lines = bufio.NewScanner(b.pty)
	if err = b.disableEcho(); err != nil {
		return nil, err
	}
	if err = b.setNarrow(); err != nil {
		return nil, err
	}
	if err = b.setupPrompt(); err != nil {
		return nil, err
	}
	return
}

// setupPrompt adjusts the prompt to the magic string, which makes it easy for
// us to identify when the prompt is ready.
func (b *Bash) setupPrompt() error {
	fmt.Fprintf(b.pty, "export PS1='\n%s\n'\n", promptMagic)
	for b.lines.Scan() {
		// log.Printf("setup read %q", b.lines.Text())
		if b.lines.Text() == promptMagic {
			break
		}
	}
	return b.lines.Err()
}

// disableEcho disables terminal echoing, which simplifies parsing by
// not having our inputs mixed into it.
func (b *Bash) disableEcho() error {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		b.pty.Fd(), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return errno
	}
	termios.Lflag &^= syscall.ECHO
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		b.pty.Fd(), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return errno
	}
	return nil
}

// setNarrow sets the pty to be very narrow, causing bash to print each
// completion on its own line.
func (b *Bash) setNarrow() error {
	ws := winsize{
		col: 2, // Note: 1 doesn't seem to have an effect.
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(b.pty.Fd()), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return errno
	}
	return nil
}

func (b *Bash) Chdir(path string) error {
	fmt.Fprintf(b.pty, "cd %q\n", path)
	// Expected output is newline + a new prompt.
	b.lines.Scan()
	if line := b.lines.Text(); line != "" {
		return fmt.Errorf("bash: unexpected line %q", line)
	}
	b.lines.Scan()
	if line := b.lines.Text(); line != promptMagic {
		return fmt.Errorf("bash: unexpected line %q", line)
	}
	return nil
}

// expand passes some input through the bash subprocess to gather
// potential expansions.
func (b *Bash) expand(input string) (expansions []string, err error) {
	// Write:
	// - the input
	// - M-? ("expand")
	// - C-u ("clear text")
	// - \n (to print a newline at the end)
	fmt.Fprintf(b.pty, "%s\x1b?\x15\n", input)

	// If there are any completions, bash prints:
	// 1) one newline (bash clearing to next line to print completions)
	// 2) list of files
	// Regardless of completions or not, it's terminated by the prompt
	// (empty line and prompt).  This means it's ambiguous if the
	// first completion matches the magic prompt string.  :(
	sawNL := false
L:
	for b.lines.Scan() {
		line := b.lines.Text()
		// log.Printf("line %q", line)
		switch {
		case line == "":
			sawNL = true
		case sawNL && line == promptMagic:
			break L
		default:
			sawNL = false
			expansions = append(expansions, line)
		}
	}
	err = b.lines.Err()
	return
}

func (b *Bash) Complete(input string) (int, []string, error) {
	expansions, err := b.expand(input)
	if err != nil {
		return 0, nil, err
	}
	if len(expansions) == 0 {
		return 0, nil, nil
	}

	// We've got some completions, but we need to guess at the correct
	// offset.  Some cases to consider:
	//   "ls " => "foo", "bar", "baz"
	//   "ls b" => "bar", "baz"
	//   "ls ./b" => "bar", "baz"
	//   "ls ./*" => "foo", "bar", "baz"
	//   "ls --c" => "--classify", "--color=", ...
	// Current logic: back up until we hit a space or a slash.
	var ofs int
	for ofs = len(input); ofs > 0; ofs-- {
		if input[ofs-1] == ' ' || input[ofs-1] == '/' {
			break
		}
	}

	return ofs, expansions, err
}
