extern crate smash;
extern crate gtk;
use smash::readline::ReadLineView;
use smash::view::Win;
use std::rc::Rc;
use std::cell::RefCell;
use gtk::prelude::*;

fn main() {
    gtk::init().unwrap();

    let win = Win::new();

    let rl = Rc::new(RefCell::new(ReadLineView::new()));
    rl.borrow_mut().rl.insert('a');

    {
        let mut win = win.borrow_mut();
        win.child = rl;
        win.gtkwin.show_all();
    }

    gtk::main();
}
