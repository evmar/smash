extern crate smash;
extern crate cairo;
extern crate gdk;
use smash::view;
use smash::readline::ReadLineView;
use smash::term::Term;

struct LogEntry {
    rl: ReadLineView,
    term: Option<Term>,
}

impl view::View for LogEntry {
    fn draw(&mut self, cr: &cairo::Context) {
        self.rl.draw(cr);
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        if let Some(ref mut term) = self.term {
            term.key(ev);
        } else {
            self.rl.key(ev);
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
            rl: ReadLineView::new(win.context.clone()),
            term: None,
        });
        win.show();
    }

    view::main();
}
