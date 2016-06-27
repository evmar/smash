extern crate cairo;
extern crate gdk;
extern crate gtk;
use gtk::prelude::*;
use std::rc::Rc;
use std::cell::RefCell;

pub trait View {
    fn draw(&mut self, cr: &cairo::Context);
    fn key(&mut self, ev: &gdk::EventKey);
}

pub struct Win {
    pub gtkwin: gtk::Window,
    pub child: Rc<RefCell<View>>,
}

impl Win {
    pub fn new(child: Rc<RefCell<View>>) -> Win {
        let win = gtk::Window::new(gtk::WindowType::Toplevel);
        win.set_default_size(400, 200);
        win.set_app_paintable(true);
        win.connect_delete_event(|_, _| {
            gtk::main_quit();
            Inhibit(false)
        });

        {
            let child = child.clone();
            win.connect_draw(move |_, cr| {
                child.borrow_mut().draw(cr);
                Inhibit(false)
            });
        }
        {
            let child = child.clone();
            let win2 = win.clone();
            win.connect_key_press_event(move |_, ev| {
                child.borrow_mut().key(ev);
                win2.queue_draw();
                Inhibit(false)
            });
        }

        Win {
            gtkwin: win,
            child: child,
        }
    }
}
