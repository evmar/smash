package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"smash/shell"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sh := shell.NewShell(cwd, map[string]string{}, nil)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("> ")
	for scanner.Scan() {
		input := scanner.Text()
		var out string
		var err error
		argv, builtin := sh.Run(input)
		if builtin != nil {
			out, err = builtin()
		} else if argv != nil {
			cmd := exec.Command(argv[0], argv[1:]...)
			cmd.Dir = sh.Cwd
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
		if out != "" {
			fmt.Fprintf(os.Stdout, "%s\n", out)
		}
		fmt.Printf("> ")
	}
}
