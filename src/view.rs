extern crate cairo;
extern crate gdk;
extern crate gtk;
use gdk::prelude::*;
use gtk::prelude::*;
use std::rc::Rc;
use std::cell::RefCell;
use std::cell::Cell;

thread_local!(static TASKS: RefCell<Vec<Box<FnMut()>>> = RefCell::new(Vec::new()));

pub fn add_task(task: Box<FnMut()>) {
    TASKS.with(|tasks| {
        tasks.borrow_mut().push(task);
    });
    gtk::idle_add(run_tasks);
}

fn run_tasks() -> gtk::Continue {
    TASKS.with(|tasks| {
        let mut tasks = tasks.borrow_mut();
        for mut t in tasks.drain(..) {
            t();
        }
    });
    Continue(false)
}

#[derive(Debug)]
#[derive(Clone)]
pub struct Layout {
    pub width: i32,
    pub height: i32,
}

impl Layout {
    pub fn add(&self, w: i32, h: i32) -> Layout {
        let layout = Layout {
            width: self.width + w,
            height: self.height + h,
        };
        if layout.width < 0 || layout.height < 0 {
            println!("layout underflow: {:?}", layout);
        }
        layout
    }
}

pub trait View {
    fn draw(&self, cr: &cairo::Context);
    fn key(&self, ev: &gdk::EventKey);
    fn layout(&self, _cr: &cairo::Context, _space: Layout) -> Layout {
        Layout {
            width: 0,
            height: 0,
        }
    }
}

pub struct NullView {}
impl View for NullView {
    fn draw(&self, _: &cairo::Context) {}
    fn key(&self, _: &gdk::EventKey) {}
}

pub struct Win {
    pub dirty_cb: Rc<Fn()>,
    pub gtkwin: gtk::Window,
    pub child: Rc<View>,
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

        let draw_pending = Rc::new(Cell::new(false));
        let dirty_cb = {
            let gtkwin = gtkwin.clone();
            let draw_pending = draw_pending.clone();
            Rc::new(move || {
                if draw_pending.get() {
                    println!("debounce dirty");
                    return;
                }
                draw_pending.set(true);
                gtkwin.queue_draw();
            })
        };

        let win = Rc::new(RefCell::new(Win {
            dirty_cb: dirty_cb,
            gtkwin: gtkwin.clone(),
            child: Rc::new(NullView {}),
        }));

        {
            let win = win.clone();
            gtkwin.connect_draw(move |_, cr| {
                let win = win.borrow();
                win.child.layout(cr,
                                 Layout {
                                     width: 600,
                                     height: 400,
                                 });
                win.child.draw(cr);
                draw_pending.set(false);
                Inhibit(false)
            });
        }
        {
            let win = win.clone();
            gtkwin.connect_key_press_event(move |_, ev| {
                let win = win.borrow();
                win.child.key(ev);
                Inhibit(false)
            });
        }

        win
    }

    pub fn create_cairo(&mut self) -> cairo::Context {
        self.gtkwin.realize();
        let gdkwin = self.gtkwin.get_window().unwrap();
        cairo::Context::create_from_window(&gdkwin)
    }

    pub fn resize(&mut self, width: i32, height: i32) {
        self.gtkwin.resize(width, height);
    }

    pub fn show(&mut self) {
        self.gtkwin.show();
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

pub fn init() {
    gtk::init().unwrap();
}

pub fn main() {
    gtk::main();
}

pub fn quit() {
    gtk::main_quit();
}
