package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"smash/shell"
)

type ShellDelegate struct {
}

func (sd *ShellDelegate) Error(error string) {
	fmt.Fprintf(os.Stderr, "%s\n", error)
}

func (sd *ShellDelegate) Start(argv []string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sd := &ShellDelegate{}
	sh := shell.NewShell(sd, cwd, map[string]string{})
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("> ")
	for scanner.Scan() {
		input := scanner.Text()
		sh.Run(sh.Parse(input))
		fmt.Printf("> ")
	}
}
