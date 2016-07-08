extern crate cairo;
extern crate gdk;
extern crate glib;
extern crate gtk;
extern crate smash;
use gdk::prelude::*;
use gtk::prelude::*;
use smash::term::Term;
use smash::view;

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

    let term = Term::new(win.borrow().context.clone(), font_extents);

    win.borrow_mut().child = Box::new(term);
    gtkwin.show_all();

    gtk::main();
}
