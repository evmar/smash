extern crate cairo;
extern crate gdk;
extern crate gtk;
use gtk::prelude::*;
use std::rc::Rc;
use std::cell::RefCell;

pub trait Context {
    fn clone_box(&self) -> Box<Context>;
    fn dirty(&mut self);
}

struct GtkContext {
    win: gtk::Window,
    draw_pending: bool,
}

impl Context for Rc<RefCell<GtkContext>> {
    fn clone_box(&self) -> Box<Context> {
        Box::new(self.clone())
    }
    fn dirty(&mut self) {
        let mut ctx = self.borrow_mut();
        println!("dirty {}", ctx.draw_pending);
        if ctx.draw_pending {
            return;
        }
        ctx.draw_pending = true;
        ctx.win.queue_draw();
    }
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

pub struct Win {
    pub context: Box<Context>,
    pub gtkwin: gtk::Window,
    pub child: Box<View>,
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

        let context = {
            let gtkwin = gtkwin.clone();
            Rc::new(RefCell::new(GtkContext {
                win: gtkwin,
                draw_pending: false,
            }))
        };

        let win = Rc::new(RefCell::new(Win {
            context: Box::new(context.clone()),
            gtkwin: gtkwin.clone(),
            child: Box::new(NullView {}),
        }));

        {
            let context = context.clone();
            let win = win.clone();
            gtkwin.connect_draw(move |_, cr| {
                let mut win = win.borrow_mut();
                win.child.draw(cr);
                context.borrow_mut().draw_pending = false;
                Inhibit(false)
            });
        }
        {
            let win = win.clone();
            gtkwin.connect_key_press_event(move |_, ev| {
                let mut win = win.borrow_mut();
                win.child.key(ev);
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
