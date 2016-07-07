extern crate cairo;
extern crate gdk;
extern crate glib;
extern crate gtk;
extern crate smash;
use gdk::prelude::*;
use gtk::prelude::*;
use smash::term::Term;
use smash::threaded_ref::ThreadedRef;
use smash::view;
use std::sync::Arc;

fn mark_dirty(context: &Arc<ThreadedRef<view::ContextRef>>) {
    // Enqueue a repaint, but put a bit of delay in it; this allows this thread
    // to do a bit more work before the paint happens.
    // TODO: ensure this actually matters in profiles.
    let context = context.clone();
    glib::timeout_add(10, move || {
        context.get().borrow_mut().dirty();
        glib::Continue(false)
    });
}

fn main() {
    gtk::init().unwrap();
    let win = view::Win::new();

    let gtkwin = win.borrow_mut().gtkwin.clone();
    gtkwin.realize();

    let font_extents = {
        let win = gtkwin.get_window().unwrap();
        let ctx = cairo::Context::create_from_window(&win);
        Term::get_font_metrics(&ctx)
    };

    gtkwin.resize(80 * font_extents.max_x_advance as i32,
                  25 * font_extents.height as i32);

    let term = {
        let context = Arc::new(ThreadedRef::new(win.borrow().context.clone()));
        Term::new(font_extents,
                  Box::new(move || {
                      mark_dirty(&context);
                  }))
    };

    win.borrow_mut().child = Box::new(term);
    gtkwin.show_all();

    gtk::main();
}
