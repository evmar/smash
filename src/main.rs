extern crate smash;

extern crate gtk;
extern crate gdk;
extern crate glib;
extern crate cairo;
use gtk::prelude::*;
use gdk::prelude::*;
use std::clone::Clone;
use std::sync::Arc;
use smash::term::Term;
use smash::view;
use smash::view::View;
use std::sync::atomic;
use std::sync::atomic::AtomicBool;
use smash::threaded_ref::ThreadedRef;

// http://stackoverflow.com/questions/31966497/
//     howto-idiomatic-rust-for-callbacks-with-gtk-rust-gnome

struct State {
    // win: gtk::Window,
    term: Term,
    dirty: Arc<AtomicBool>,
}

impl View for State {
    fn draw(&mut self, cr: &cairo::Context) {
        self.term.draw(cr);
        self.dirty.store(false, atomic::Ordering::SeqCst);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.term.key(ev);
    }
}

fn mark_dirty(dirty: &Arc<AtomicBool>, win: ThreadedRef<gtk::Window>) {
    let was_dirty = dirty.compare_and_swap(false, true, atomic::Ordering::SeqCst);
    if was_dirty {
        return;
    }

    // Enqueue a repaint, but put a bit of delay in it; this allows this thread
    // to do a bit more work before the paint happens.
    // TODO: ensure this actually matters in profiles.
    glib::timeout_add(10, move || {
        win.get().queue_draw();
        glib::Continue(false)
    });
}

fn wmain() {
    let win = view::Win::new();

    let gtkwin = win.borrow_mut().gtkwin.clone();
    gtkwin.realize();

    let font_extents = {
        match gtkwin.get_window() {
            Some(ref win) => {
                let ctx = cairo::Context::create_from_window(&win);
                Term::get_font_metrics(&ctx)
            }
            None => panic!("no window"),
        }
    };

    gtkwin.resize(80 * font_extents.max_x_advance as i32,
                  25 * font_extents.height as i32);

    let dirty = Arc::new(AtomicBool::new(false));
    let dirty2 = ThreadedRef::new(gtkwin.clone());
    let term = {
        let dirty = dirty.clone();
        Term::new(font_extents,
                  Box::new(move || {
                      mark_dirty(&dirty, dirty2.clone());
                  }))
    };
    let state = State {
        // win: win.clone(),
        dirty: dirty,
        term: term,
    };

    win.borrow_mut().child = Box::new(state);
    gtkwin.show_all();

    gtk::main();
}

fn main() {
    gtk::init().unwrap();
    wmain();
}
