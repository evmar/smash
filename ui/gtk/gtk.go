package gtk

/*
#cgo pkg-config: gtk+-3.0
#cgo LDFLAGS: -L${SRCDIR} -lsmashgtk
#include <gtk/gtk.h>
#include "smashgtk.h"
*/
import "C"
import (
	"smash/keys"
	"smash/ui"
	"unsafe"

	"github.com/martine/gocairo/cairo"
)

type UI struct{}

type Window struct {
	win unsafe.Pointer
	// Store the delegate here so this interface isn't gc'd.
	delegate ui.WinDelegate
}

func Init() *UI {
	C.smash_gtk_init()
	return &UI{}
}

func (ui *UI) Enqueue(f func()) {
	panic("x")
}

func (ui *UI) NewWindow(delegate ui.WinDelegate) ui.Win {
	win := &Window{
		delegate: delegate,
	}
	win.win = C.smash_gtk_new_window(unsafe.Pointer(&win.delegate))
	return win
}

func (ui *UI) Loop(win ui.Win) {
	C.gtk_main()
}

func (ui *UI) Quit() {
	C.gtk_main_quit()
}

func (w *Window) Dirty() {
	C.gtk_widget_queue_draw((*C.GtkWidget)(w.win))
}

//export callDraw
func callDraw(delegateP unsafe.Pointer, crP unsafe.Pointer) {
	delegate := (*ui.WinDelegate)(delegateP)
	cr := cairo.BorrowContext(crP)
	(*delegate).Draw(cr)
}

//export callKey
func callKey(delegateP unsafe.Pointer, keyP unsafe.Pointer) {
	delegate := (*ui.WinDelegate)(delegateP)
	gkey := (*C.GdkEventKey)(keyP)

	switch gkey.keyval {
	case C.GDK_KEY_Shift_L, C.GDK_KEY_Shift_R,
		C.GDK_KEY_Control_L, C.GDK_KEY_Control_R,
		C.GDK_KEY_Alt_L, C.GDK_KEY_Alt_R,
		C.GDK_KEY_Meta_L, C.GDK_KEY_Meta_R,
		C.GDK_KEY_Super_L, C.GDK_KEY_Super_R:
		return
	}
	key := keys.Key{}
	key.Sym = keys.Sym(gkey.keyval)
	if gkey.state&C.GDK_CONTROL_MASK != 0 {
		key.Mods |= keys.ModControl
	}
	if gkey.state&C.GDK_MOD1_MASK != 0 {
		key.Mods |= keys.ModMeta
	}
	(*delegate).Key(key)
}
