extern crate cairo;
extern crate smash;
extern crate gdk;
use smash::readline::ReadLineView;
use smash::view;
use smash::view::View;
use smash::view::Win;

struct Padding {
    child: Box<View>,
}

impl View for Padding {
    fn draw(&mut self, cr: &cairo::Context) {
        cr.translate(20.0, 20.0);
        self.child.draw(cr);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.child.key(ev);
    }
}

fn main() {
    view::init();

    let win = Win::new();
    {
        let mut win = win.borrow_mut();

        let rl = ReadLineView::new(win.context.clone());
        rl.borrow_mut().rl.insert("a");

        let padding = Padding { child: Box::new(rl) };

        win.child = Box::new(padding);
        win.show();
    }

    view::main();
}
