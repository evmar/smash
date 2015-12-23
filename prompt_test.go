package main

import (
	"smash/readline"
	"smash/shell"
	"testing"
)

type testPromptDelegate struct {
}

func (tpd *testPromptDelegate) OnPromptAccept(string) bool {
	panic("x")
}
func (tpd *testPromptDelegate) GetPromptAbsolutePosition(pv *PromptView) (int, int) {
	return 0, 0
}
func (tpd *testPromptDelegate) Complete(input string) (int, []string, error) {
	if input == "ls l" {
		return 3, []string{"log", "logview.go"}, nil
	}
	panic("x")
}

func TestComplete(t *testing.T) {
	parent := NewTestViewHost()
	delegate := &testPromptDelegate{}
	config := &readline.Config{}
	shell := shell.NewShell(".", map[string]string{}, delegate)
	pv := NewPromptView(parent, delegate, config, shell)
	pv.readline.Text = []byte("ls l")
	pv.readline.Pos = 4
	pv.StartComplete()
	parent.runQueue(true)
}
