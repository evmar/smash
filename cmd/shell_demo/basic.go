package main

import (
	"bufio"
	"fmt"
	"os"
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
		cmd, builtin := sh.Run(input)
		if cmd != nil {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
		} else if builtin != nil {
			out, err = builtin()
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
