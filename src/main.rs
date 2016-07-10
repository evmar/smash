extern crate smash;
extern crate cairo;
extern crate gdk;
use smash::view;
use smash::readline::ReadLineView;
use smash::term::Term;
use std::rc::Rc;
use std::cell::RefCell;

struct Prompt {
    rl: Rc<RefCell<ReadLineView>>,
}

impl view::View for Prompt {
    fn draw(&mut self, cr: &cairo::Context) {
        cr.translate(10.0, 10.0);
        self.rl.draw(cr);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.rl.key(ev);
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
            prompt: Prompt { rl: ReadLineView::new(win.context.clone()) },
            term: None,
        });
        win.show();
    }

    view::main();
}
