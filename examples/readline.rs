extern crate cairo;
extern crate smash;
extern crate gtk;
extern crate gdk;
use smash::readline::ReadLineView;
use smash::view::View;
use smash::view::Win;
use gtk::prelude::*;

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
    gtk::init().unwrap();

    let win = Win::new();

    let mut rl = ReadLineView::new(win.borrow().context.clone());
    rl.rl.insert("a");

    let padding = Padding { child: Box::new(rl) };

    {
        let mut win = win.borrow_mut();
        win.child = Box::new(padding);
        win.gtkwin.show_all();
    }

    gtk::main();
}
