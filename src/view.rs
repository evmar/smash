extern crate cairo;
extern crate gdk;
extern crate gtk;
use gtk::prelude::*;
use std::rc::Rc;
use std::cell::RefCell;

pub struct Context {
    pub dirty: bool,
}

pub trait View {
    fn draw(&mut self, cr: &cairo::Context);
    fn key(&mut self, ev: &gdk::EventKey);
}

pub struct NullView {}
impl View for NullView {
    fn draw(&mut self, _: &cairo::Context) {}
    fn key(&mut self, _: &gdk::EventKey) {}
}

pub type ContextRef = Rc<RefCell<Context>>;

pub struct Win {
    pub gtkwin: gtk::Window,
    pub child: Box<View>,
    pub context: ContextRef,
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
            context: Rc::new(RefCell::new(Context { dirty: false })),
            gtkwin: gtkwin.clone(),
            child: Box::new(NullView {}),
        }));

        {
            let win = win.clone();
            gtkwin.connect_draw(move |_, cr| {
                let mut win = win.borrow_mut();
                win.child.draw(cr);
                if win.context.borrow().dirty {
                    win.gtkwin.queue_draw();
                }
                Inhibit(false)
            });
        }
        {
            let win = win.clone();
            gtkwin.connect_key_press_event(move |_, ev| {
                let mut win = win.borrow_mut();
                win.child.key(ev);
                if win.context.borrow().dirty {
                    win.gtkwin.queue_draw();
                }
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
