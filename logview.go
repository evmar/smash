package main

import (
	"os/exec"
	"strings"
	"time"

	"github.com/martine/gocairo/cairo"

	"smash/bash"
	"smash/keys"
	"smash/readline"
	"smash/shell"
	"smash/ui"
)

type LogEntry struct {
	prompt *PromptView
	term   *TermView
}

func (le *LogEntry) Height() int {
	h := le.prompt.Height()
	if le.term != nil {
		h += le.term.Height()
	}
	return h
}

type LogView struct {
	ViewBase
	Entries []*LogEntry
	shell   *shell.Shell

	rlconfig     *readline.Config
	scrollOffset int
	scrollAnim   *ui.Lerp
}

func NewLogView(parent View) (*LogView, error) {
	lv := &LogView{
		ViewBase: ViewBase{parent},
		rlconfig: readline.NewConfig(),
	}
	cwd := ""
	var env map[string]string
	bash, err := bash.StartBash()
	if err != nil {
		return nil, err
	}
	lv.shell = shell.NewShell(cwd, env, bash)
	lv.addEntry()
	return lv, nil
}

func (lv *LogView) addEntry() {
	e := &LogEntry{
		prompt: NewPromptView(lv, lv, lv.rlconfig, lv.shell),
	}
	lv.Entries = append(lv.Entries, e)
}

func ParseCommand(input string) *exec.Cmd {
	// TODO: something correct.
	args := strings.Split(input, " ")
	return exec.Command(args[0], args[1:]...)
}

func (lv *LogView) OnPromptAccept(input string) bool {
	e := lv.Entries[len(lv.Entries)-1]
	e.term = NewTermView(lv)
	cmd, builtin := lv.shell.Run(input)
	e.term.OnExit = func() {
		lv.addEntry()
	}
	if cmd != nil {
		e.term.Start(cmd)
	} else if builtin != nil {
		// TODO: async.
		output, err := builtin()
		if err != nil {
			e.term.term.DisplayString(err.Error())
		} else {
			e.term.term.DisplayString(output)
		}
		e.term.Finish()
	}
	return true
}

func (lv *LogView) GetPromptAbsolutePosition(pv *PromptView) (int, int) {
	x, y := lv.GetWindow().win.GetContentPosition()
	for _, e := range lv.Entries {
		y += e.Height()
	}
	y -= lv.scrollOffset
	return x, y
}

func (l *LogView) Draw(cr *cairo.Context) {
	cr.Translate(0, float64(-l.scrollOffset))
	y := 0
	for _, e := range l.Entries {
		e.prompt.Draw(cr)
		h := e.prompt.Height()
		y += h
		cr.Translate(0, float64(h))
		if e.term != nil {
			e.term.Draw(cr)
			h = e.term.Height()
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
				l.scrollAnim = ui.NewLerp(&l.scrollOffset, scrollOffset, 40*time.Millisecond)
				l.GetWindow().win.AddAnimation(l.scrollAnim)
			} else {
				// TODO adjust existing anim
			}
		}
	}
}

func (l *LogView) Key(key keys.Key) {
	e := l.Entries[len(l.Entries)-1]
	if e.term != nil && e.term.Running {
		e.term.Key(key)
	} else {
		e.prompt.Key(key)
	}
}

func (l *LogView) Scroll(dy int) {
}
