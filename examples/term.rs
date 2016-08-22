extern crate smash;
use smash::term::Term;
use smash::view;
use std::rc::Rc;

fn main() {
    view::init();
    let win = view::Win::new();

    {
        let font_extents = {
            let ctx = win.create_cairo();
            Term::get_font_metrics(&ctx)
        };

        win.resize(80 * font_extents.max_x_advance as i32,
                   25 * font_extents.height as i32);

        let mut term = Term::new(win.dirty_cb.clone(), font_extents);
        term.spawn(&["bash"], Box::new(|| {}));
        *win.child.borrow_mut() = Rc::new(term);

        win.show();
    }

    view::main();
}
