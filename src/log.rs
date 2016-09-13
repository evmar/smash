extern crate cairo;
extern crate gdk;
use std::rc::Rc;
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
    layout: Layout,
}

impl LogEntry {
    pub fn new(id: usize, dirty: Rc<Fn()>, accept: Box<FnMut(usize, &str)>) -> LogEntry {
        let mut le = LogEntry {
            id: id,
            prompt: Prompt::new(dirty),
            term: None,
            layout: Layout::new(),
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

    fn key(&mut self, ev: &gdk::EventKey) {
        if let Some(ref mut term) = self.term {
            term.key(ev);
        } else {
            self.prompt.key(ev);
        }
    }

    fn relayout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        let mut layout = self.prompt.relayout(cr, space);
        if let Some(ref mut term) = self.term {
            let tlayout = term.relayout(cr,
                                        Layout {
                                            width: space.width,
                                            height: space.height - layout.height,
                                        });
            layout = layout.add(tlayout.width, tlayout.height);
        }
        self.layout = layout;
        layout
    }
    fn get_layout(&self) -> Layout {
        self.layout
    }
}

pub struct Log {
    shell: Rc<RefCell<shell::Shell>>,
    entries: Vec<LogEntry>,
    dirty: Rc<Fn()>,
    font_extents: cairo::FontExtents,
    scroll_offset: i32,
    layout: Layout,
}

impl Log {
    pub fn new(dirty: Rc<Fn()>, font_extents: &cairo::FontExtents) -> Rc<RefCell<Log>> {
        let log = Rc::new(RefCell::new(Log {
            shell: Rc::new(RefCell::new(shell::Shell::new())),
            entries: Vec::new(),
            dirty: dirty,
            font_extents: font_extents.clone(),
            scroll_offset: 0,
            layout: Layout::new(),
        }));
        Log::new_entry(&log);
        log
    }

    pub fn new_entry(rlog: &Rc<RefCell<Log>>) {
        let mut log = rlog.borrow_mut();
        let id = log.entries.len();
        let accept_cb = {
            // accept is only called once; hack around missing Box<FnOnce>.
            let mut once = Some(rlog.clone());
            Box::new(move |id: usize, str: &str| {
                let log = once.take().unwrap();
                let cmd = shell::parse(str);
                view::add_task(move || {
                    Log::start(log, id, cmd);
                })
            })
        };
        let entry = LogEntry::new(id, log.dirty.clone(), accept_cb);
        log.entries.push(entry);
        (log.dirty)();
    }

    fn start(rlog: Rc<RefCell<Log>>, id: usize, cmd: shell::Command) {
        let done = {
            let rlog = rlog.clone();
            Box::new(move || {
                Log::new_entry(&rlog);
            })
        };

        let mut log = rlog.borrow_mut();
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
        log.entries[id].term = Some(term);
    }
}

impl view::View for Log {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        cr.save();
        let mut y = 0;
        cr.translate(0.0, -self.scroll_offset as f64);
        for (i, entry) in self.entries.iter().enumerate() {
            let height = entry.get_layout().height;
            y += height;
            if y < self.scroll_offset {
                cr.translate(0.0, height as f64);
                continue;
            }
            let last = i == self.entries.len() - 1;
            entry.draw(cr, focus && last);
            cr.translate(0.0, height as f64);
        }
        cr.restore();
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        let last = self.entries.len() - 1;
        self.entries[last].key(ev);
    }
    fn relayout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        let mut height = 0;
        for ref mut entry in self.entries.iter_mut() {
            let entry_layout = entry.relayout(cr, space);
            height += entry_layout.height;
        }
        if height > space.height {
            self.scroll_offset = height - space.height;
        }
        // Ignore the computed height, because the log fills all vertical
        // space given to it.
        self.layout = space;
        self.layout
    }
    fn get_layout(&self) -> Layout {
        self.layout
    }
}
