extern crate cairo;
extern crate gdk;
use std::rc::Rc;
use std::cell::Cell;
use std::cell::RefCell;
use prompt::Prompt;
use shell::{Command, Shell};
use term::Term;
use view;
use view::Layout;

pub struct LogEntry {
    shell: Rc<Shell>,
    prompt: Prompt,
    term: RefCell<Option<Term>>,
    layout: Cell<Layout>,
}

impl LogEntry {
    pub fn new(shell: Rc<Shell>,
               dirty: Rc<Fn()>,
               font_extents: cairo::FontExtents,
               done: Box<Fn()>)
               -> Rc<LogEntry> {
        let le = Rc::new(LogEntry {
            shell: shell,
            prompt: Prompt::new(dirty.clone()),
            term: RefCell::new(None),
            layout: Cell::new(Layout::new()),
        });

        let accept_cb = {
            // The accept callback from readline can potentially be
            // called multiple times, but we only want create a
            // terminal once.  Capture all the needed state in a
            // moveable temporary.
            let le = le.clone();
            let mut once = Some((le.clone(), dirty, font_extents, done));
            Box::new(move |str: &str| {
                if let Some(once) = once.take() {
                    let cmd = le.shell.parse(str);
                    view::add_task(move || {
                        match cmd {
                            Command::Builtin(_) => {}
                            Command::External(argv) => {
                                let argv: Vec<_> = argv.iter().map(|s| s.as_str()).collect();
                                let (le, dirty, font_extents, done) = once;
                                *le.term.borrow_mut() =
                                    Some(Term::new(dirty, font_extents, argv.as_slice(), done));
                            }
                        }
                    })
                }
            })
        };
        le.prompt.set_accept_cb(accept_cb);
        le
    }
}

impl view::View for LogEntry {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        if let Some(ref term) = *self.term.borrow() {
            self.prompt.draw(cr, false);
            cr.save();
            let height = self.prompt.get_layout().height as f64;
            cr.translate(0.0, height);
            term.draw(cr, focus);
            cr.restore();
        } else {
            self.prompt.draw(cr, focus);
        }
    }

    fn key(&self, ev: &gdk::EventKey) {
        if let Some(ref term) = *self.term.borrow() {
            term.key(ev);
        } else {
            self.prompt.key(ev);
        }
    }

    fn relayout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        let mut layout = self.prompt.relayout(cr, space);
        if let Some(ref term) = *self.term.borrow() {
            let tlayout = term.relayout(cr,
                                        Layout {
                                            width: space.width,
                                            height: space.height - layout.height,
                                        });
            layout = layout.add(tlayout.width, tlayout.height);
        }
        self.layout.set(layout);
        layout
    }
    fn get_layout(&self) -> Layout {
        self.layout.get()
    }
}

pub struct Log {
    shell: Rc<Shell>,
    entries: RefCell<Vec<Rc<LogEntry>>>,
    dirty: Rc<Fn()>,
    font_extents: cairo::FontExtents,
    scroll_offset: Cell<i32>,
    layout: Cell<Layout>,
}

impl Log {
    pub fn new(dirty: Rc<Fn()>, font_extents: &cairo::FontExtents) -> Rc<Log> {
        let log = Rc::new(Log {
            shell: Rc::new(Shell::new()),
            entries: RefCell::new(Vec::new()),
            dirty: dirty,
            font_extents: font_extents.clone(),
            scroll_offset: Cell::new(0),
            layout: Cell::new(Layout::new()),
        });
        Log::new_entry(&log);
        log
    }

    pub fn new_entry(log: &Rc<Log>) {
        let entry = {
            let log = log.clone();
            LogEntry::new(log.shell.clone(),
                          log.dirty.clone(),
                          log.font_extents,
                          Box::new(move || {
                              Log::new_entry(&log);
                          }))
        };
        log.entries.borrow_mut().push(entry);
    }
}

impl view::View for Log {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        let entries = self.entries.borrow();
        cr.save();
        let mut y = 0;
        let scroll_offset = self.scroll_offset.get();
        cr.translate(0.0, -scroll_offset as f64);
        for (i, entry) in entries.iter().enumerate() {
            let height = entry.get_layout().height;
            y += height;
            if y < scroll_offset {
                cr.translate(0.0, height as f64);
                continue;
            }
            let last = i == entries.len() - 1;
            entry.draw(cr, focus && last);
            cr.translate(0.0, height as f64);
        }
        cr.restore();
    }
    fn key(&self, ev: &gdk::EventKey) {
        let entries = self.entries.borrow();
        entries[entries.len() - 1].key(ev);
    }
    fn relayout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        let entries = self.entries.borrow();
        let mut height = 0;
        for entry in &*entries {
            let entry_layout = entry.relayout(cr, space);
            height += entry_layout.height;
        }
        if height > space.height {
            self.scroll_offset.set(height - space.height);
        }
        // Ignore the computed height, because the log fills all vertical
        // space given to it.
        self.layout.set(space);
        self.layout.get()
    }
    fn get_layout(&self) -> Layout {
        self.layout.get()
    }
}
