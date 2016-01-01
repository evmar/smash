package main

import (
	"os/exec"
	"time"

	"github.com/martine/gocairo/cairo"

	"github.com/martine/smash/bash"
	"github.com/martine/smash/keys"
	"github.com/martine/smash/readline"
	"github.com/martine/smash/shell"
	"github.com/martine/smash/ui"
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

	height int

	shell *shell.Shell

	rlconfig     *readline.Config
	scrollOffset int
	scrollAnim   *ui.Lerp
}

func NewLogView(parent View, height int) (*LogView, error) {
	lv := &LogView{
		ViewBase: ViewBase{parent},
		rlconfig: readline.NewConfig(),
		height:   height,
	}
	cwd := ""
	bash, err := bash.StartBash()
	if err != nil {
		return nil, err
	}
	lv.shell = shell.NewShell(cwd, nil, bash)
	lv.shell.LoadEnv()
	lv.addEntry()
	return lv, nil
}

func (lv *LogView) addEntry() {
	e := &LogEntry{
		prompt: NewPromptView(lv, lv, lv.rlconfig, lv.shell),
	}
	lv.Entries = append(lv.Entries, e)
}

func (lv *LogView) OnPromptAccept(input string) bool {
	e := lv.Entries[len(lv.Entries)-1]
	argv, builtin := lv.shell.Run(input)
	if argv == nil && builtin == nil {
		// Empty input.
		lv.addEntry()
		return false // Don't add to history.
	}

	e.term = NewTermView(lv)
	e.term.OnExit = func() {
		lv.addEntry()
	}
	if builtin != nil {
		// TODO: async.
		output, err := builtin()
		if err != nil {
			if err, ok := err.(shell.Exit); ok {
				lv.GetWindow().ui.Quit()
			} else {
				e.term.term.DisplayString(err.Error())
			}
		} else {
			e.term.term.DisplayString(output)
		}
		e.term.Finish()
	} else if argv != nil {
		// TODO: async.
		path, err := lv.shell.LookPath(argv[0])
		if err != nil {
			e.term.term.DisplayString(err.Error())
			e.term.Finish()
			return true
		}
		cmd := &exec.Cmd{
			Path: path,
			Args: argv,
		}
		cmd.Dir = lv.shell.Cwd
		e.term.Start(cmd)
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
		h := e.prompt.Height()
		if y+h >= l.scrollOffset {
			e.prompt.Draw(cr)
		}
		y += h
		cr.Translate(0, float64(h))
		if e.term != nil {
			h = e.term.Height()
			if y+h >= l.scrollOffset {
				e.term.Draw(cr)
			}
			y += h
			cr.Translate(0, float64(h))
		}
	}
	if y > l.height {
		scrollOffset := y - l.height
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

func (l *LogView) Key(key keys.Key) bool {
	e := l.Entries[len(l.Entries)-1]
	if e.term != nil && e.term.Running {
		return e.term.Key(key)
	} else {
		return e.prompt.Key(key)
	}
}

func (l *LogView) Scroll(dy int) {
}
