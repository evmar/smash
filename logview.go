package main

import (
	"smash/keys"

	"github.com/martine/gocairo/cairo"
)

type LogEntry struct {
	pb *PromptBuf
	tb *TermBuf
}

type LogView struct {
	ViewBase
	Entries []*LogEntry
}

func NewLogView(parent View) *LogView {
	lv := &LogView{
		ViewBase: ViewBase{parent},
	}
	e := &LogEntry{
		pb: NewPromptBuf(lv, lv.Accept),
	}
	e.pb.Accept = lv.Accept
	lv.Entries = append(lv.Entries, e)
	return lv
}

func (l *LogView) Accept(input string) {
	e := l.Entries[len(l.Entries)-1]
	tb := NewTermBuf(l)
	e.tb = tb
	go tb.runBash()
}

func (l *LogView) Draw(cr *cairo.Context) {
	for _, e := range l.Entries {
		e.pb.Draw(cr)
		if e.tb != nil {
			e.tb.Draw(cr)
		}
	}
}

func (l *LogView) Key(key keys.Key) {
	e := l.Entries[len(l.Entries)-1]
	e.pb.Key(key)
}

func (l *LogView) Scroll(dy int) {
}
