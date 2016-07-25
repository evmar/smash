package smash

import (
	"github.com/evmar/smash/keys"
	"github.com/evmar/smash/ui"

	"github.com/evmar/gocairo/cairo"
)

type View interface {
	GetWindow() *Window
	Draw(cr *cairo.Context)
	Key(key keys.Key) bool
	Scroll(dy int)
	Dirty()
	Enqueue(f func())
}

type ViewBase struct {
	Parent View
}

func (vb *ViewBase) GetWindow() *Window {
	return vb.Parent.GetWindow()
}

func (vb *ViewBase) Dirty() {
	vb.Parent.Dirty()
}

func (vb *ViewBase) Enqueue(f func()) {
	vb.Parent.Enqueue(f)
}

type Window struct {
	ui  ui.UI
	win ui.Win

	View View
	font *Font
}

func NewWindow(ui ui.UI) *Window {
	win := &Window{ui: ui, font: NewMonoFont()}
	win.win = ui.NewWindow(win, true)
	return win
}

func (win *Window) GetWindow() *Window {
	return win
}

func (win *Window) GetUiWindow() ui.Win {
	return win.win
}

func (win *Window) Mapped() {
	panic("x")
}

func (w *Window) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(1, 1, 1)
	cr.Paint()
	w.View.Draw(cr)
}

func (w *Window) Key(key keys.Key) bool {
	return w.View.Key(key)
}

func (w *Window) Scroll(dy int) {
	w.View.Scroll(dy)
}

func (w *Window) Dirty() {
	w.win.Dirty()
}

func (w *Window) Enqueue(f func()) {
	w.ui.Enqueue(f)
}

func (w *Window) Show() {
	w.win.Show()
}
