extern crate smash;
extern crate gtk;
use smash::readline::ReadLineView;
use std::rc::Rc;
use std::cell::RefCell;
use gtk::prelude::*;

fn main() {
    gtk::init().unwrap();

    let win = gtk::Window::new(gtk::WindowType::Toplevel);
    win.set_default_size(400, 200);
    win.set_app_paintable(true);
    win.connect_delete_event(|_, _| {
        gtk::main_quit();
        Inhibit(false)
    });

    let rl = Rc::new(RefCell::new(ReadLineView::new()));
    rl.borrow_mut().rl.insert('a');

    {
        let rl = rl.clone();
        win.connect_draw(move |_, cr| {
            rl.borrow_mut().draw(cr);
            Inhibit(false)
        });
    }
    {
        let rl = rl.clone();
        let win2 = win.clone();
        win.connect_key_press_event(move |_, ev| {
            rl.borrow_mut().key(ev);
            win2.queue_draw();
            Inhibit(false)
        });
    }

    win.show_all();

    gtk::main();
}
