extern crate smash;
extern crate cairo;
extern crate gdk;
use smash::log;
use smash::view;
use smash::term::Term;
use std::rc::Rc;


fn main() {
    view::init();
    let win = view::Win::new();

    {
        let mut win = win.borrow_mut();
        let font_extents = {
            let ctx = win.create_cairo();
            Term::get_font_metrics(&ctx)
        };
        win.child = log::Log::new(win.context.clone(), &font_extents);
        win.show();
    }

    view::main();
}
