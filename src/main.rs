extern crate smash;
extern crate cairo;
extern crate gdk;
use smash::log;
use smash::view;
use smash::term::Term;


fn main() {
    view::init();
    let rwin = view::Win::new();
    {
        let mut win = rwin.borrow_mut();
        win.resize(600, 400);

        let font_extents = {
            let ctx = win.create_cairo();
            Term::get_font_metrics(&ctx)
        };
        win.child = log::Log::new(win.dirty_cb.clone(), &font_extents);
        win.show();
    }

    view::main();
}
