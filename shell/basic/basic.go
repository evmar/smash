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
	sh := shell.NewShell(cwd, map[string]string{})
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("> ")
	for scanner.Scan() {
		input := scanner.Text()
		cmd, err := sh.Run(sh.Parse(input))
		if cmd != nil {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
		fmt.Printf("> ")
	}
}
