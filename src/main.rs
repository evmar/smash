extern crate cairo;
extern crate gdk;
extern crate glib;
extern crate gtk;
extern crate smash;
use gtk::prelude::*;
use smash::term::Term;
use smash::view;

fn main() {
    gtk::init().unwrap();
    let win = view::Win::new();

    let font_extents = {
        let ctx = win.borrow_mut().create_cairo();
        Term::get_font_metrics(&ctx)
    };

    let gtkwin = &win.borrow_mut().gtkwin;
    gtkwin.resize(80 * font_extents.max_x_advance as i32,
                  25 * font_extents.height as i32);

    let term = Term::new(win.borrow().context.clone(), font_extents);

    win.borrow_mut().child = Box::new(term);
    gtkwin.show_all();

    gtk::main();
}
