package gtk

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include "smashgtk.h"
*/
import "C"
import (
	"smash/base"
	"smash/keys"
	"smash/ui"
	"time"
	"unsafe"

	"github.com/martine/gocairo/cairo"
)

type UI struct {
	// Functions enqueued to do work on the main goroutine.
	enqueued chan func()
}

type Window struct {
	gtkWin unsafe.Pointer
	// Store the delegate here so this interface isn't gc'd.
	delegate ui.WinDelegate

	anims map[base.Anim]bool
}

func Init() *UI {
	C.smash_gtk_init()
	return &UI{
		enqueued: make(chan func(), 1),
	}
}

//export callIdle
func callIdle(data unsafe.Pointer) int {
	ui := (*UI)(data)
	for {
		select {
		case f := <-ui.enqueued:
			f()
		default:
			return 0 // Don't run again
		}
	}
}

func (ui *UI) Enqueue(f func()) {
	// Proxy the function to the main thread by using g_idle_add.
	// You'd be tempted to want to pass f to g_idle_add directly, but
	// (a) passing closures into C code is annoying, and (b) we need a
	// reference to the function on the Go side for GC reasons.  So
	// it's easier to just put the function in a channel that pulls it
	// back out on the other thread.
	ui.enqueued <- f
	C.g_idle_add(C.GSourceFunc(C.smash_idle_cb), C.gpointer(ui))
}

func (ui *UI) NewWindow(delegate ui.WinDelegate) ui.Win {
	win := &Window{
		delegate: delegate,
		anims:    make(map[base.Anim]bool),
	}
	win.gtkWin = C.smash_gtk_new_window(unsafe.Pointer(&win.delegate))
	return win
}

func (ui *UI) Loop(win ui.Win) {
	C.gtk_main()
}

func (ui *UI) Quit() {
	C.gtk_main_quit()
}

func (w *Window) Dirty() {
	C.gtk_widget_queue_draw((*C.GtkWidget)(w.gtkWin))
}

//export callTick
func callTick(data unsafe.Pointer) bool {
	win := (*Window)(data)
	// TODO: use gdk_frame_clock_get_frame_time here instead of Go time.
	now := time.Now()
	for anim := range win.anims {
		if !anim.Frame(now) {
			delete(win.anims, anim)
		}
	}
	return len(win.anims) > 0
}

func (w *Window) AddAnimation(anim base.Anim) {
	if len(w.anims) == 0 {
		C.smash_start_ticks(unsafe.Pointer(w), (*C.GtkWidget)(w.gtkWin))
	}
	w.anims[anim] = true
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

	rune := C.gdk_keyval_to_unicode(gkey.keyval)
	key := keys.Key{}
	key.Sym = keys.Sym(rune)
	if gkey.state&C.GDK_CONTROL_MASK != 0 {
		key.Mods |= keys.ModControl
	}
	if gkey.state&C.GDK_MOD1_MASK != 0 {
		key.Mods |= keys.ModMeta
	}
	(*delegate).Key(key)
}
