package main

import (
	"smash/keys"
	"smash/readline"
	"smash/shell"
	"testing"

	"github.com/martine/gocairo/cairo"
)

type testView struct {
	win Window

	uiQueue chan func()
}

func NewTestView() *testView {
	return &testView{
		uiQueue: make(chan func()),
	}
}

func (tv *testView) runQueue(wait bool) {
	if wait {
		f := <-tv.uiQueue
		f()
	}
	for {
		select {
		case f := <-tv.uiQueue:
			f()
		default:
			return
		}
	}
}

func (tv *testView) GetWindow() *Window {
	return &tv.win
}
func (tv *testView) Draw(cr *cairo.Context) {}
func (tv *testView) Key(key keys.Key) bool {
	return false
}
func (tv *testView) Scroll(dy int)    {}
func (tv *testView) Dirty()           {}
func (tv *testView) Enqueue(f func()) {}

type testPromptDelegate struct {
}

func (tpd *testPromptDelegate) OnPromptAccept(string) bool {
	panic("x")
}
func (tpd *testPromptDelegate) GetPromptAbsolutePosition(pv *PromptView) (int, int) {
	panic("x")
}

func TestComplete(t *testing.T) {
	parent := NewTestView()
	delegate := &testPromptDelegate{}
	config := &readline.Config{}
	shell := &shell.Shell{}
	pv := NewPromptView(parent, delegate, config, shell)
	pv.readline.Text = []byte("ls l")
	pv.readline.Pos = 4
	pv.StartComplete()
	parent.runQueue(true)
}
