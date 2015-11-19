package main

import (
	"os/exec"
	"smash/base"
	"smash/keys"
	"smash/readline"
	"strings"
	"time"

	"github.com/martine/gocairo/cairo"
)

type LogEntry struct {
	pb *PromptBuf
	tb *TermView
}

type LogView struct {
	ViewBase
	Entries []*LogEntry

	rlconfig     *readline.Config
	scrollOffset int
	scrollAnim   *base.Lerp
}

func NewLogView(parent View) *LogView {
	lv := &LogView{
		ViewBase: ViewBase{parent},
		rlconfig: readline.NewConfig(),
	}
	lv.addEntry()
	return lv
}

func (lv *LogView) addEntry() {
	e := &LogEntry{
		pb: NewPromptBuf(lv, lv.rlconfig, lv.Accept),
	}
	lv.Entries = append(lv.Entries, e)
}

func ParseCommand(input string) *exec.Cmd {
	// TODO: something correct.
	args := strings.Split(input, " ")
	return exec.Command(args[0], args[1:]...)
}

func (l *LogView) Accept(input string) bool {
	e := l.Entries[len(l.Entries)-1]
	tb := NewTermView(l)
	e.tb = tb
	e.tb.OnExit = func() {
		l.addEntry()
	}
	tb.Start(ParseCommand(input))
	return true
}

func (l *LogView) Draw(cr *cairo.Context) {
	cr.Translate(0, float64(-l.scrollOffset))
	y := 0
	for _, e := range l.Entries {
		e.pb.Draw(cr)
		h := e.pb.Height()
		y += h
		cr.Translate(0, float64(h))
		if e.tb != nil {
			e.tb.Draw(cr)
			h = e.tb.Height()
			y += h
			cr.Translate(0, float64(h))
		}
	}
	if y > 400 {
		scrollOffset := y - 400
		if l.scrollOffset != scrollOffset {
			if l.scrollAnim != nil && l.scrollAnim.Done {
				l.scrollAnim = nil
			}
			if l.scrollAnim == nil {
				l.scrollAnim = base.NewLerp(&l.scrollOffset, scrollOffset, 40*time.Millisecond)
				l.GetWindow().AddAnimation(l.scrollAnim)
			} else {
				// TODO adjust existing anim
			}
		}
	}
}

func (l *LogView) Key(key keys.Key) {
	e := l.Entries[len(l.Entries)-1]
	if e.tb != nil && e.tb.Running {
		e.tb.Key(key)
	} else {
		e.pb.Key(key)
	}
}

func (l *LogView) Scroll(dy int) {
}
