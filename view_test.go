package main

import (
	"smash/keys"

	"github.com/martine/gocairo/cairo"
)

type testViewHost struct {
	win Window

	uiQueue chan func()
}

func NewTestViewHost() *testViewHost {
	font := NewMonoFont()
	font.fakeMetrics()
	return &testViewHost{
		win: Window{
			font: font,
		},
		uiQueue: make(chan func(), 10),
	}
}

func (tv *testViewHost) runQueue(wait bool) {
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

func (tv *testViewHost) GetWindow() *Window {
	return &tv.win
}
func (tv *testViewHost) Draw(cr *cairo.Context) {}
func (tv *testViewHost) Key(key keys.Key) bool {
	return false
}
func (tv *testViewHost) Scroll(dy int) {}
func (tv *testViewHost) Dirty()        {}
func (tv *testViewHost) Enqueue(f func()) {
	tv.uiQueue <- f
}
