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
    term: RefCell<Option<Term>>,
}

impl LogEntry {
    pub fn new(dirty: Rc<Fn()>,
               font_extents: &cairo::FontExtents,
               done: Box<Fn()>)
               -> Rc<LogEntry> {
        let le = Rc::new(LogEntry {
            prompt: Prompt::new(ReadLineView::new(dirty.clone())),
            term: RefCell::new(None),
        });

        let accept_cb = {
            // The accept callback from readline can potentially be
            // called multiple times, but we only want create a
            // terminal once.  Capture all the needed state in a
            // moveable temporary.
            let mut once = Some((le.clone(), dirty, font_extents.clone(), done));
            Box::new(move |str: &str| {
                if let Some(once) = once.take() {
                    let text = String::from(str);
                    view::add_task(move || {
                        let (le, dirty, font_extents, done) = once;
                        *le.term.borrow_mut() =
                            Some(Term::new(dirty, font_extents, &[&text], done));
                    })
                }
            })
        };
        le.prompt.rl.rl.borrow_mut().accept_cb = accept_cb;
        le
    }
}

impl view::View for LogEntry {
    fn draw(&self, cr: &cairo::Context) {
        self.prompt.draw(cr);
        if let Some(ref term) = *self.term.borrow() {
            cr.save();
            let height = self.prompt.height.get() as f64;
            cr.translate(0.0, height);
            term.draw(cr);
            cr.restore();
        }
    }
    fn key(&self, ev: &gdk::EventKey) {
        if let Some(ref term) = *self.term.borrow() {
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
    entries: Vec<Rc<LogEntry>>,
    done: bool,
}

impl Log {
    pub fn new(dirty: Rc<Fn()>, font_extents: &cairo::FontExtents) -> Rc<RefCell<Log>> {
        let log = Rc::new(RefCell::new(Log {
            entries: Vec::new(),
            done: false,
        }));
        let e = {
            let log = log.clone();
            LogEntry::new(dirty,
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
