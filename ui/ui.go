package ui

import (
	"github.com/evmar/smash/keys"

	"github.com/evmar/gocairo/cairo"
)

type Win interface {
	Dirty()

	GetCairo() *cairo.Context
	SetSize(width, height int)
	SetPosition(x, y int)
	GetContentPosition() (int, int)
	Show()
	Close()

	AddAnimation(anim Anim)
}

type WinDelegate interface {
	// Mapped is called when the window is first shown.
	Mapped()
	// Draw draws the display content into the backing store.
	Draw(cr *cairo.Context)
	// Key is called when there's a keypress on the window.
	Key(key keys.Key) bool
	// Scrolled is called when there's a scroll event.
	Scroll(dy int)
}

type UI interface {
	NewWindow(delegate WinDelegate, toplevel bool) Win
	Enqueue(f func())
	Loop()
	Quit()
}
