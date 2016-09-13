extern crate cairo;
extern crate gdk;
use std::rc::Rc;
use std::cell::Cell;
use std::cell::RefCell;
use prompt::Prompt;
use shell;
use term::Term;
use view;
use view::Layout;

pub struct LogEntry {
    id: usize,
    prompt: Prompt,
    term: Option<Term>,
    layout: Cell<Layout>,
}

impl LogEntry {
    pub fn new(id: usize, dirty: Rc<Fn()>, accept: Box<FnMut(usize, &str)>) -> LogEntry {
        let le = LogEntry {
            id: id,
            prompt: Prompt::new(dirty),
            term: None,
            layout: Cell::new(Layout::new()),
        };
        {
            // The accept callback from readline can potentially be
            // called multiple times, but we only want to accept once.
            let mut once = Some((id, accept));
            le.prompt.set_accept_cb(Box::new(move |str: &str| {
                if let Some((id, mut accept)) = once.take() {
                    accept(id, str);
                }
            }));
        }
        le
    }
}

impl view::View for LogEntry {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        if let Some(ref term) = self.term {
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
        if let Some(ref term) = self.term {
            term.key(ev);
        } else {
            self.prompt.key(ev);
        }
    }

    fn relayout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        let mut layout = self.prompt.relayout(cr, space);
        if let Some(ref term) = self.term {
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
    shell: Rc<RefCell<shell::Shell>>,
    entries: RefCell<Vec<LogEntry>>,
    dirty: Rc<Fn()>,
    font_extents: cairo::FontExtents,
    scroll_offset: Cell<i32>,
    layout: Cell<Layout>,
}

impl Log {
    pub fn new(dirty: Rc<Fn()>, font_extents: &cairo::FontExtents) -> Rc<Log> {
        let log = Rc::new(Log {
            shell: Rc::new(RefCell::new(shell::Shell::new())),
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
        let id = log.entries.borrow().len();
        let accept_cb = {
            // accept is only called once; hack around missing Box<FnOnce>.
            let mut once = Some(log.clone());
            Box::new(move |id: usize, str: &str| {
                let log = once.take().unwrap();
                let cmd = shell::parse(str);
                view::add_task(move || {
                    Log::start(log, id, cmd);
                })
            })
        };
        let entry = LogEntry::new(id, log.dirty.clone(), accept_cb);
        log.entries.borrow_mut().push(entry);
        (log.dirty)();
    }

    fn start(log: Rc<Log>, id: usize, cmd: shell::Command) {
        let done = {
            let log = log.clone();
            Box::new(move || {
                Log::new_entry(&log);
            })
        };

        let mut term = Term::new(log.dirty.clone(), log.font_extents);
        match cmd {
            shell::Command::Builtin(f) => {
                f(&mut log.shell.borrow_mut());
                term.cleanup();
                done();
            }
            shell::Command::External(argv) => {
                let argv: Vec<_> = argv.iter().map(|s| s.as_str()).collect();
                term.spawn(argv.as_slice(), done);
            }
        }
        log.entries.borrow_mut()[id].term = Some(term);
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
