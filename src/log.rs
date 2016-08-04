extern crate cairo;
extern crate gdk;
use std::rc::Rc;
use std::cell::Cell;
use std::cell::RefCell;
use readline::ReadLineView;
use term::Term;
use view;
use view::View;
use view::Layout;

struct Prompt {
    rl: Rc<ReadLineView>,
    height: Cell<i32>,
}

impl Prompt {
    fn new(rl: Rc<ReadLineView>) -> Prompt {
        Prompt {
            rl: rl,
            height: Cell::new(0),
        }
    }
}

impl view::View for Prompt {
    fn draw(&self, cr: &cairo::Context) {
        cr.save();
        cr.set_source_rgb(0.7, 0.7, 0.7);
        cr.new_path();
        cr.move_to(5.0, 8.0);
        let height = self.height.get() as f64;
        cr.line_to(13.0, height / 2.0);
        cr.line_to(5.0, height - 8.0);
        cr.fill();

        cr.translate(18.0, 5.0);
        self.rl.draw(cr);
        cr.restore();
    }
    fn key(&self, ev: &gdk::EventKey) {
        self.rl.key(ev);
    }

    fn layout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        let rlsize = self.rl.layout(cr, space.add(-20, -10));
        let layout = rlsize.add(20, 10);
        self.height.set(layout.height);
        layout
    }
}

pub struct LogEntry {
    prompt: Prompt,
    term: Option<RefCell<Term>>,
}

impl LogEntry {
    pub fn new(context: view::ContextRef,
               font_extents: &cairo::FontExtents,
               done: Box<Fn()>)
               -> LogEntry {
        LogEntry {
            prompt: Prompt::new(ReadLineView::new(context.clone())),
            term: Some(RefCell::new(Term::new(context.clone(), *font_extents, &["ls"], done))),
        }
    }
}

impl view::View for LogEntry {
    fn draw(&self, cr: &cairo::Context) {
        self.prompt.draw(cr);
        if let Some(ref term) = self.term {
            cr.save();
            let height = self.prompt.height.get() as f64;
            cr.translate(0.0, height);
            term.draw(cr);
            cr.restore();
        }
    }
    fn key(&self, ev: &gdk::EventKey) {
        if let Some(ref term) = self.term {
            term.key(ev);
        } else {
            self.prompt.key(ev);
        }
    }
    fn layout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        self.prompt.layout(cr, space.clone());
        space
    }
}

pub struct Log {
    entries: Vec<LogEntry>,
    done: bool,
}

impl Log {
    pub fn new(context: view::ContextRef, font_extents: &cairo::FontExtents) -> Rc<RefCell<Log>> {
        let log = Rc::new(RefCell::new(Log {
            entries: Vec::new(),
            done: false,
        }));
        let e = {
            let log = log.clone();
            LogEntry::new(context,
                          font_extents,
                          Box::new(move || {
                              log.borrow_mut().done = true;
                              println!("done");
                          }))
        };
        log.borrow_mut().entries.push(e);
        log
    }
}

impl view::View for RefCell<Log> {
    fn draw(&self, cr: &cairo::Context) {
        let entries = &self.borrow().entries;
        entries[0].draw(cr)
    }
    fn key(&self, ev: &gdk::EventKey) {
        let entries = &self.borrow().entries;
        entries[0].key(ev)
    }
    fn layout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        let entries = &self.borrow().entries;
        entries[0].layout(cr, space)
    }
}
