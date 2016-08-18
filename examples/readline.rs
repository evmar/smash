extern crate cairo;
extern crate smash;
extern crate gdk;
use smash::readline::ReadLineView;
use smash::view;
use smash::view::{View, Layout};
use smash::view::Win;
use std::rc::Rc;

struct Padding {
    child: Rc<View>,
}

impl View for Padding {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        cr.translate(20.0, 20.0);
        self.child.draw(cr, focus);
    }
    fn key(&self, ev: &gdk::EventKey) {
        self.child.key(ev);
    }
    fn get_layout(&self) -> Layout {
        self.child.get_layout().add(40, 40)
    }
}

fn main() {
    view::init();

    let win = Win::new();
    {
        let rl = ReadLineView::new(win.dirty_cb.clone());
        let padding = Padding { child: rl.clone() };

        *win.child.borrow_mut() = Rc::new(padding);
        win.show();
    }

    view::main();
}
