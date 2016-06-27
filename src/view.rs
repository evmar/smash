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

pub struct NullView {}
impl View for NullView {
    fn draw(&mut self, _: &cairo::Context) {}
    fn key(&mut self, _: &gdk::EventKey) {}
}

pub struct Win {
    pub gtkwin: gtk::Window,
    pub child: Rc<RefCell<View>>,
}

impl Win {
    pub fn new() -> Rc<RefCell<Win>> {
        let gtkwin = gtk::Window::new(gtk::WindowType::Toplevel);
        gtkwin.set_default_size(400, 200);
        gtkwin.set_app_paintable(true);
        gtkwin.connect_delete_event(|_, _| {
            gtk::main_quit();
            Inhibit(false)
        });

        let win = Rc::new(RefCell::new(Win {
            gtkwin: gtkwin.clone(),
            child: Rc::new(RefCell::new(NullView {})),
        }));

        {
            let win = win.clone();
            gtkwin.connect_draw(move |_, cr| {
                let win = win.borrow_mut();
                win.child.borrow_mut().draw(cr);
                Inhibit(false)
            });
        }
        {
            let win = win.clone();
            gtkwin.connect_key_press_event(move |_, ev| {
                let win = win.borrow_mut();
                win.child.borrow_mut().key(ev);
                win.gtkwin.queue_draw();
                Inhibit(false)
            });
        }

        win
    }
}
