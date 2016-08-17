extern crate smash;
extern crate cairo;
extern crate gdk;
use smash::log;
use smash::view;
use smash::term::Term;


fn main() {
    view::init();
    let win = view::Win::new();

    {
        let font_extents = {
            let ctx = win.create_cairo();
            Term::get_font_metrics(&ctx)
        };
        *win.child.borrow_mut() = log::Log::new(win.dirty_cb.clone(), &font_extents);
        win.show();
    }

    view::main();
}
