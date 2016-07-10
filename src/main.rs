extern crate smash;
extern crate cairo;
extern crate gdk;
use smash::view;
use smash::view::Layout;
use smash::readline::ReadLineView;
use smash::term::Term;
use std::rc::Rc;
use std::cell::RefCell;

struct Prompt {
    rl: Rc<RefCell<ReadLineView>>,
    height: i32,
}

impl Prompt {
    fn new(rl: Rc<RefCell<ReadLineView>>) -> Prompt {
        Prompt {
            rl: rl,
            height: 0,
        }
    }
}

impl view::View for Prompt {
    fn draw(&mut self, cr: &cairo::Context) {
        cr.set_source_rgb(0.7, 0.7, 0.7);
        cr.new_path();
        cr.move_to(5.0, 8.0);
        cr.line_to(13.0, self.height as f64 / 2.0);
        cr.line_to(5.0, self.height as f64 - 8.0);
        cr.fill();

        cr.translate(18.0, 5.0);
        self.rl.draw(cr);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.rl.key(ev);
    }

    fn layout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        let rlsize = self.rl.layout(cr, space.add(-20, -10));
        let layout = rlsize.add(20, 10);
        self.height = layout.height;
        layout
    }
}

struct LogEntry {
    prompt: Prompt,
    term: Option<Term>,
}

impl view::View for LogEntry {
    fn draw(&mut self, cr: &cairo::Context) {
        self.prompt.draw(cr);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        if let Some(ref mut term) = self.term {
            term.key(ev);
        } else {
            self.prompt.key(ev);
        }
    }
    fn layout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        self.prompt.layout(cr, space)
    }
}

// struct Log {
//     entries: Vec<LogEntry>,
// }

fn main() {
    view::init();
    let win = view::Win::new();

    {
        let mut win = win.borrow_mut();
        win.child = Box::new(LogEntry {
            prompt: Prompt::new(ReadLineView::new(win.context.clone())),
            term: None,
        });
        win.show();
    }

    view::main();
}
