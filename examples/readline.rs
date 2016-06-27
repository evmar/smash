extern crate cairo;
extern crate smash;
extern crate gtk;
extern crate gdk;
use smash::readline::ReadLineView;
use smash::view::View;
use smash::view::Win;
use std::rc::Rc;
use std::cell::RefCell;
use gtk::prelude::*;

struct Padding {
    child: Rc<RefCell<View>>,
}
impl View for Padding {
    fn draw(&mut self, cr: &cairo::Context) {
        cr.translate(20.0, 20.0);
        self.child.borrow_mut().draw(cr);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.child.borrow_mut().key(ev);
    }
}

fn main() {
    gtk::init().unwrap();

    let win = Win::new();

    let rl = Rc::new(RefCell::new(ReadLineView::new()));
    rl.borrow_mut().rl.insert('a');

    let padding = Padding { child: rl };

    {
        let mut win = win.borrow_mut();
        win.child = Rc::new(RefCell::new(padding));
        win.gtkwin.show_all();
    }

    gtk::main();
}
