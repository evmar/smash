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

pub fn is_modifier_key_event(ev: &gdk::EventKey) -> bool {
    match ev.get_keyval() {
        gdk::enums::key::Caps_Lock |
        gdk::enums::key::Control_L |
        gdk::enums::key::Control_R |
        gdk::enums::key::Shift_L |
        gdk::enums::key::Shift_R |
        gdk::enums::key::Alt_L |
        gdk::enums::key::Alt_R |
        gdk::enums::key::Meta_L |
        gdk::enums::key::Meta_R => true,
        _ => false,
    }
}
