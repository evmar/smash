package main

import (
	"github.com/martine/smash"
	"github.com/martine/smash/readline"
	"github.com/martine/smash/ui/gtk"
)

type del struct{}

func (d *del) OnPromptAccept(string) bool {
	return true
}

func (d *del) GetPromptAbsolutePosition(pv *smash.PromptView) (int, int) {
	return 0, 0
}

func main() {
	ui := gtk.Init()

	win := smash.NewWindow(ui)

	prompt := smash.NewPromptView(win, &del{}, readline.NewConfig(), nil)
	win.View = prompt

	win.GetUiWindow().SetSize(600, 100)

	win.Show()
	ui.Loop()
}
