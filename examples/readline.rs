extern crate cairo;
extern crate smash;
extern crate gdk;
use smash::readline::ReadLineView;
use smash::view;
use smash::view::{View, Layout};
use smash::view::Win;
use std::rc::Rc;
use std::cell::RefCell;

struct Padding {
    child: Rc<RefCell<View>>,
}

impl View for Padding {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        cr.translate(20.0, 20.0);
        self.child.borrow().draw(cr, focus);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.child.borrow_mut().key(ev);
    }
    fn get_layout(&self) -> Layout {
        self.child.borrow().get_layout().add(40, 40)
    }
}

fn main() {
    view::init();

    let rwin = Win::new();
    {
        let mut win = rwin.borrow_mut();
        let rl = ReadLineView::new(win.dirty_cb.clone());
        let padding = Padding { child: rl.clone() };

        win.child = Rc::new(RefCell::new(padding));
        win.show();
    }

    view::main();
}
