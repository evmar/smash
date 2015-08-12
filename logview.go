package main

import (
	"os/exec"
	"smash/keys"
	"strings"

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
	lv.addEntry()
	return lv
}

func (lv *LogView) addEntry() {
	e := &LogEntry{
		pb: NewPromptBuf(lv, lv.Accept),
	}
	lv.Entries = append(lv.Entries, e)
}

func ParseCommand(input string) *exec.Cmd {
	// TODO: something correct.
	args := strings.Split(input, " ")
	return exec.Command(args[0], args[1:]...)
}

func (l *LogView) Accept(input string) {
	e := l.Entries[len(l.Entries)-1]
	tb := NewTermBuf(l)
	e.tb = tb
	e.tb.OnExit = func() {
		l.addEntry()
	}
	tb.Start(ParseCommand(input))
}

func (l *LogView) Draw(cr *cairo.Context) {
	for _, e := range l.Entries {
		e.pb.Draw(cr)
		cr.Translate(0, float64(e.pb.Height()))
		if e.tb != nil {
			e.tb.Draw(cr)
			cr.Translate(0, float64(e.tb.Height()))
		}
	}
}

func (l *LogView) Key(key keys.Key) {
	e := l.Entries[len(l.Entries)-1]
	e.pb.Key(key)
}

func (l *LogView) Scroll(dy int) {
}
