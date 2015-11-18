package ui

import (
	"smash/keys"

	"github.com/martine/gocairo/cairo"
)

type Win interface {
	Dirty()
}

type WinDelegate interface {
	// Mapped is called when the window is first shown.
	Mapped()
	// Draw draws the display content into the backing store.
	Draw(cr *cairo.Context)
	// Key is called when there's a keypress on the window.
	Key(key keys.Key)
	// Scrolled is called when there's a scroll event.
	Scroll(dy int)
}

type UI interface {
	NewWindow(delegate WinDelegate) Win
	Enqueue(f func())
	Loop(win Win)
	Quit()
}
