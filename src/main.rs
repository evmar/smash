mod term;
mod vt100;
mod pty;
mod byte_scanner;

extern crate gtk;
extern crate gdk;
extern crate glib;
extern crate cairo;
use gtk::prelude::*;
use gdk::prelude::*;
use std::cell::RefCell;
use std::clone::Clone;
use std::rc::Rc;
use std::sync::Arc;
use term::Term;
use std::sync::atomic;
use std::sync::atomic::AtomicBool;

// This file is full of nasty hacks because the Rust GTK bindings are
// still a work in progress.  In particular, callbacks are tricky (because
// they must have static lifetime), and getting callbacks across threads
// is impossible without using thread-local storage.

// http://stackoverflow.com/questions/31966497/howto-idiomatic-rust-for-callbacks-with-gtk-rust-gnome

thread_local!(
    static TLS_WIN: RefCell<Option<gtk::Window>> = RefCell::new(None)
);

struct State {
    // win: gtk::Window,
    term: Term,
    dirty: Arc<AtomicBool>,
}

impl State {
    fn draw(&mut self, cr: &cairo::Context) {
        self.term.draw(cr);
    }
}

fn mark_dirty(dirty: &Arc<AtomicBool>) {
    let was_dirty = dirty.compare_and_swap(false, true, atomic::Ordering::SeqCst);
    if was_dirty {
        return;
    }

    // Enqueue a repaint, but put a bit of delay in it; this allows this thread
    // to do a bit more work before the paint happens.
    // TODO: ensure this actually matters in profiles.
    glib::timeout_add(10, || {
        TLS_WIN.with(|w| {
            if let Some(ref w) = *w.borrow() {
                w.queue_draw();
            }
        });
        glib::Continue(false)
    });
}

fn wmain() {
    let win = gtk::Window::new(gtk::WindowType::Toplevel);
    TLS_WIN.with(|w| *w.borrow_mut() = Some(win.clone()));

    win.realize();

    let font_extents = {
        match win.get_window() {
            Some(ref win) => {
                let ctx = cairo::Context::create_from_window(&win);
                Term::get_font_metrics(&ctx)
            }
            None => panic!("no window"),
        }
    };

    win.resize(80 * font_extents.max_x_advance as i32,
               25 * font_extents.height as i32);

    let dirty = Arc::new(AtomicBool::new(false));
    let term = {
        let dirty = dirty.clone();
        Term::new(font_extents, Box::new(move || mark_dirty(&dirty)))
    };
    let state = Rc::new(RefCell::new(State {
        // win: win.clone(),
        dirty: dirty,
        term: term,
    }));

    win.set_app_paintable(true);
    win.connect_delete_event(|_, _| {
        gtk::main_quit();
        Inhibit(false)
    });

    {
        let state = state.clone();
        win.connect_draw(move |_, cr| {
            let mut state = state.borrow_mut();
            state.draw(cr);
            state.dirty.store(false, atomic::Ordering::SeqCst);
            Inhibit(false)
        });
    }

    {
        let state = state.clone();
        win.connect_key_press_event(move |_, ev| {
            state.borrow_mut().term.key(ev);
            Inhibit(false)
        });
    }

    win.show_all();

    gtk::main();
}

fn main() {
    gtk::init().unwrap();
    wmain();
}
