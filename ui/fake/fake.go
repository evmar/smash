package fake

import (
	"github.com/martine/smash/ui"

	"github.com/martine/gocairo/cairo"
)

type UI struct {
	uiQueue chan func()
}

func NewUI() *UI {
	return &UI{
		uiQueue: make(chan func(), 10),
	}
}

func (fui *UI) RunQueue(wait bool) {
	if wait {
		f := <-fui.uiQueue
		f()
	}
	for {
		select {
		case f := <-fui.uiQueue:
			f()
		default:
			return
		}
	}
}

func (fui *UI) NewWindow(delegate ui.WinDelegate, toplevel bool) ui.Win {
	return &Win{}
}

func (fui *UI) Enqueue(f func()) {
	fui.uiQueue <- f
}

func (fui *UI) Loop() {
}

func (fui *UI) Quit() {
}

type Win struct {
	ui *UI
}

func (w *Win) Dirty()                         {}
func (w *Win) GetCairo() *cairo.Context       { return nil }
func (w *Win) SetSize(width, height int)      {}
func (w *Win) SetPosition(x, y int)           {}
func (w *Win) GetContentPosition() (int, int) { return 0, 0 }
func (w *Win) Show()                          {}
func (w *Win) Close()                         {}
func (w *Win) AddAnimation(anim ui.Anim)      {}
