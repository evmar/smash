// Package xlib wraps enough of xlib to produce Cairo drawing surfaces.
package xlib

/*
#cgo pkg-config: x11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
*/
import "C"

import (
	"bytes"
	"log"
	"syscall"
	"unsafe"

	"github.com/martine/gocairo/cairo"

	"smash/base"
)

const EINTR syscall.Errno = 4

type Display struct {
	dpy *C.Display

	// eventReady has a token when X events are available to be read.
	eventReady chan bool
	// drawReady has a token when Go code desires a redraw.
	drawReady chan bool
	// quit has a token when it's time to shutdown.
	quit chan bool

	anims *base.AnimSet
}

type WinDelegate interface {
	// Mapped is called when the window is first shown.
	Mapped()
	// Draw draws the display content into the backing store.
	Draw(cr *cairo.Context)
	// Key is called when there's a keypress on the window.
	Key(key *base.Key)
	// Scrolled is called when there's a scroll event.
	Scroll(dy int)
}

type Window struct {
	delegate WinDelegate

	dpy *Display
	xw  C.Window

	// winSurface is a Cairo surface pointed at the window.
	winSurface *cairo.XlibSurface
	// bufSurface is a Cairo surface pointed at the backing store
	// double buffer.
	bufSurface *cairo.XlibSurface

	width, height int
}

// Resize handles a resize of the window.
func (win *Window) resize(w, h int) {
	if win.width == w && win.height == h {
		return
	}
	win.width, win.height = w, h

	if win.winSurface == nil {
		visual := C.XDefaultVisual(win.dpy.dpy, 0)
		win.winSurface = cairo.XlibSurfaceCreate(
			unsafe.Pointer(win.dpy.dpy), uint64(win.xw), unsafe.Pointer(visual),
			w, h)
	} else {
		win.winSurface.SetSize(w, h)
	}
	win.bufSurface = &cairo.XlibSurface{win.winSurface.CreateSimilar(cairo.ContentColor, w, h)}

	cr := cairo.Create(win.bufSurface.Surface)
	win.delegate.Draw(cr)
	win.repaint(0, 0, w, h)
}

var flip bool

// Repaint paints the backing store onto the X window.
func (win *Window) repaint(x, y, w, h int) {
	cr := cairo.Create(win.winSurface.Surface)
	cr.Rectangle(float64(x), float64(y), float64(w), float64(h))
	cr.Clip()
	if flip {
		cr.SetSourceRGB(1, 0, 0)
	} else {
		cr.SetSourceSurface(win.bufSurface.Surface, 0, 0)
		// TODO clip to exposed area?
		cr.SetOperator(cairo.OperatorSource)
	}
	// flip = !flip
	cr.Paint()
}

func (win *Window) Dirty() {
	select {
	case win.dpy.drawReady <- true:
	default:
	}
}

// processXEvents does one read from the X fd, then processes all pending
// events.
func (dpy *Display) processXEvents(win *Window) {
	var event C.XEvent
	events := C.XEventsQueued(dpy.dpy, C.QueuedAfterReading)
	for i := 0; i < int(events); i++ {
		C.XNextEvent(dpy.dpy, &event)
		e := &event
		typ := xEventType(*(*C.int)(unsafe.Pointer(e)))
		// log.Printf("typ %s", typ.String())
		switch typ {
		case C.ConfigureNotify:
			e := (*C.XConfigureEvent)(unsafe.Pointer(e))
			win.resize(int(e.width), int(e.height))
		case C.Expose:
			e := (*C.XExposeEvent)(unsafe.Pointer(e))
			win.repaint(int(e.x), int(e.y), int(e.width), int(e.height))
		case C.KeyPress:
			e := (*C.XKeyEvent)(unsafe.Pointer(e))

			key := &base.Key{}
			var buf [8]byte
			var keysym C.KeySym
			C.XLookupString(e, (*C.char)(unsafe.Pointer(&buf)), 8, &keysym, nil)
			nulpos := bytes.Index(buf[:], []byte{0})
			if nulpos > 0 {
				if e.state&C.Mod1Mask != 0 {
					// Alt: send an escape before the next key.
					key.Text += "\x1b"
				}
				key.Text += string(buf[:nulpos])
			} else {
				switch keysym {
				case C.XK_Left:
					key.Special = base.KeyLeft
				case C.XK_Right:
					key.Special = base.KeyRight
				case C.XK_Up:
					key.Special = base.KeyUp
				case C.XK_Down:
					key.Special = base.KeyDown
				case C.XK_Shift_L, C.XK_Shift_R:
				case C.XK_Control_L, C.XK_Control_R:
				case C.XK_Meta_L, C.XK_Meta_R:
				case C.XK_Alt_L, C.XK_Alt_R:
				case C.XK_Super_L, C.XK_Super_R:
				case C.XK_Hyper_L, C.XK_Hyper_R:
				default:
					log.Printf("unhandled keysym %#v: %#v", keysym, e)
				}
			}
			win.delegate.Key(key)
		case C.KeyRelease:
			// ignore
		case C.ButtonPress:
			e := (*C.XButtonEvent)(unsafe.Pointer(e))
			switch e.button {
			case 4, 5:
				dy := -1
				if e.button == 5 {
					dy = 1
				}
				win.delegate.Scroll(dy)
			case 6, 7:
				// TODO horizontal scroll.
			default:
				log.Printf("unhandled button %#v", e)
			}
		case C.MapNotify:
			win.delegate.Mapped()
		case C.ReparentNotify:
			// ignore
		default:
			if typ > C.GenericEvent {
				// Cairo triggers shm events, which show up as extension
				// events with greater ids than C.GenericEvent.
				continue
			}
			log.Printf("unhandled ev %s", typ.String())
		}
	}
	C.XFlush(dpy.dpy)
}

// waitUntilReadable blocks until fd is readable.
func waitUntilReadable(fd int) error {
	var fds syscall.FdSet
	fds.Bits[0] = 1 << uint(fd)
	for {
		_, err := syscall.Select(fd+1, &fds, nil, nil, nil)
		if err != nil {
			if err == EINTR {
				continue
			}
			return err
		}
		return nil // readable
	}
}

// Loop runs the main X loop.
func (dpy *Display) Loop(win *Window) {
	awaitEvent := make(chan bool)
	go func() {
		xfd := int(C.XConnectionNumber(dpy.dpy))
		for {
			<-awaitEvent
			err := waitUntilReadable(xfd)
			if err != nil {
				log.Fatalf("select %#v", err)
			}
			dpy.eventReady <- true
		}
	}()

	C.XFlush(dpy.dpy)
	awaitEvent <- true
	draw := false
	for {
		nextFrame := dpy.anims.NextFrame(draw)
		select {
		case <-dpy.eventReady:
			dpy.processXEvents(win)
			awaitEvent <- true
		case <-dpy.drawReady:
			draw = true
		case t := <-nextFrame:
			dpy.anims.Run()
			// log.Printf("draw %s", t)
			t = t
			cr := cairo.Create(win.bufSurface.Surface)
			win.delegate.Draw(cr)
			win.repaint(0, 0, win.width, win.height)
			C.XSync(dpy.dpy, 0)
			draw = false
		case <-dpy.quit:
			return
		}
	}
}

func (dpy *Display) Quit() {
	select {
	case dpy.quit <- true:
	default:
	}
}

func OpenDisplay(anims *base.AnimSet) *Display {
	return &Display{
		dpy:        C.XOpenDisplay(nil),
		eventReady: make(chan bool),
		drawReady:  make(chan bool, 1),
		quit:       make(chan bool, 1),
		anims:      anims,
	}
}

func (d *Display) NewWindow(delegate WinDelegate) *Window {
	w := C.XCreateSimpleWindow(d.dpy, C.XDefaultRootWindow(d.dpy),
		0, 0, 640, 388,
		0, 0, C.XWhitePixel(d.dpy, 0))
	C.XSelectInput(d.dpy, w, C.StructureNotifyMask|C.SubstructureNotifyMask|C.ExposureMask|C.KeyPress|C.KeyRelease|C.ButtonPress)
	C.XMapWindow(d.dpy, w)

	return &Window{
		delegate: delegate,
		dpy:      d,
		xw:       w,
	}
}
