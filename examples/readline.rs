extern crate cairo;
extern crate smash;
extern crate gdk;
use smash::readline::ReadLineView;
use smash::view;
use smash::view::View;
use smash::view::Win;
use std::rc::Rc;

struct Padding {
    child: Rc<View>,
}

impl View for Padding {
    fn draw(&self, cr: &cairo::Context) {
        cr.translate(20.0, 20.0);
        self.child.draw(cr);
    }
    fn key(&self, ev: &gdk::EventKey) {
        self.child.key(ev);
    }
}

fn main() {
    view::init();

    let win = Win::new();
    {
        let mut win = win.borrow_mut();

        let rl = ReadLineView::new(win.context.clone());
        let padding = Padding { child: rl.clone() };

        win.child = Rc::new(padding);
        win.show();
    }

    view::main();
}
