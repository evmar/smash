extern crate cairo;
extern crate gdk;
extern crate glib;

use gtk::prelude::*;
use std::fs;
use std::io::Write;
use std::os::unix::io::AsRawFd;
use std::os::unix::io::FromRawFd;
use std::str;
use std::sync::Arc;
use std::sync::Mutex;
use std::sync::atomic::AtomicBool;
use std::sync::atomic;
use std::thread;
use std::time;
use std::sync::mpsc;

use pty;
use threaded_ref::ThreadedRef;
use view::View;
use view;
use vt100;

type Color = (u8, u8, u8);
const DEFAULT_BG: Color = (0xf7, 0xf7, 0xf7);
const DEFAULT_FG: Color = (0, 0, 0);
const ANSI_COLORS: [Color; 8] = [(0x2e, 0x34, 0x36),
                                 (0xcc, 0x00, 0x00),
                                 (0x4e, 0x9a, 0x06),
                                 (0xc4, 0xa0, 0x00),
                                 (0x34, 0x65, 0xa4),
                                 (0x75, 0x50, 0x7b),
                                 (0x06, 0x98, 0x9a),
                                 (0xd3, 0xd7, 0xcf)];

const ANSI_BRIGHT_COLORS: [Color; 8] = [(0x55, 0x57, 0x53),
                                        (0xef, 0x29, 0x29),
                                        (0x8a, 0xe2, 0x34),
                                        (0xfc, 0xe9, 0x4f),
                                        (0x72, 0x9f, 0xcf),
                                        (0xad, 0x7f, 0xa8),
                                        (0x34, 0xe2, 0xe2),
                                        (0xee, 0xee, 0xec)];

fn use_color(cr: &cairo::Context, color: Color) {
    cr.set_source_rgb((color.0 as f64) / 255.0,
                      (color.1 as f64) / 255.0,
                      (color.2 as f64) / 255.0);
}

fn duration_in_ms(dur: time::Duration) -> u64 {
    dur.as_secs() * 1000 + (dur.subsec_nanos() as u64 / 1000000)
}

pub struct Term {
    pub font_metrics: cairo::FontExtents,
    vt: Arc<Mutex<vt100::VT>>,
    stdin: mpsc::Sender<Box<[u8]>>,
    draw_pending: Arc<AtomicBool>,
    last_paint: time::Instant,
}

impl Term {
    pub fn new(context: view::ContextRef,
               font_extents: cairo::FontExtents,
               command: &[&str])
               -> Term {
        let rf = pty::spawn(command);
        pty::set_size(&rf, 25, 80);
        let (stdin_send, stdin_recv) = mpsc::channel();

        {
            // Spawn a thread that forwards data from the channel into
            // the pty.
            let mut stdin = unsafe { fs::File::from_raw_fd(rf.as_raw_fd()) };
            thread::spawn(move || {
                loop {
                    let buf: Box<[u8]> = stdin_recv.recv().unwrap();
                    if stdin.write(&*buf).is_err() {
                        break;
                    }
                }
            });
        }

        let term = Term {
            font_metrics: font_extents,
            vt: Arc::new(Mutex::new(vt100::VT::new())),
            stdin: stdin_send.clone(),
            draw_pending: Arc::new(AtomicBool::new(false)),
            last_paint: time::Instant::now(),
        };

        {
            let mut rf = rf;
            let vt = term.vt.clone();
            let draw_pending = term.draw_pending.clone();
            let context = Arc::new(ThreadedRef::new(context.clone()));
            thread::spawn(move || {
                let mut r = vt100::VTReader::new(&*vt, stdin_send);
                while r.read(&mut rf) {
                    let draw_was_pending =
                        draw_pending.compare_and_swap(false, true, atomic::Ordering::SeqCst);
                    if draw_was_pending {
                        continue;
                    }

                    // Enqueue a repaint, but put a bit of delay in it; this allows this thread
                    // to do a bit more work before the paint happens.
                    // TODO: ensure this actually matters in profiles.
                    let context = context.clone();
                    glib::timeout_add(10, move || {
                        context.get().borrow_mut().dirty();
                        glib::Continue(false)
                    });
                }
            })
        };

        term
    }

    pub fn get_font_metrics(cr: &cairo::Context) -> cairo::FontExtents {
        cr.select_font_face("mono",
                            cairo::enums::FontSlant::Normal,
                            cairo::enums::FontWeight::Normal);
        cr.set_font_size(15.0);
        cr.font_extents()
    }

    fn use_font(&mut self, cr: &cairo::Context) {
        cr.select_font_face("mono",
                            cairo::enums::FontSlant::Normal,
                            cairo::enums::FontWeight::Normal);
        cr.set_font_size(15.0);
        if self.font_metrics.height == 0.0 {
            let m = cr.font_extents();
            // println!("metrics {:?} {:?}", m.ascent, m.descent);
            self.font_metrics = m;
        }
    }

    fn draw_span(&self, cr: &cairo::Context, x: f64, attr: vt100::Attr, text: &str) {
        cr.select_font_face("mono",
                            cairo::enums::FontSlant::Normal,
                            if attr.bold() {
                                cairo::enums::FontWeight::Bold
                            } else {
                                cairo::enums::FontWeight::Normal
                            });

        let (fg, bg) = {
            let bg = attr.bg().map_or(DEFAULT_BG, |c| ANSI_COLORS[c]);
            let fg = attr.fg().map_or(DEFAULT_FG, |c| {
                if attr.bold() {
                    ANSI_BRIGHT_COLORS[c]
                } else {
                    ANSI_COLORS[c]
                }
            });
            if attr.inverse() {
                (bg, fg)
            } else {
                (fg, bg)
            }
        };

        if bg != DEFAULT_BG {
            use_color(cr, bg);
            cr.rectangle(x,
                         0.0,
                         text.len() as f64 * self.font_metrics.max_x_advance,
                         self.font_metrics.height);
            cr.fill();
        }

        cr.move_to(x, self.font_metrics.ascent);
        use_color(cr, fg);
        cr.show_text(text);
    }
}

impl View for Term {
    fn draw(&mut self, cr: &cairo::Context) {
        self.draw_pending.store(false, atomic::Ordering::SeqCst);
        let now = time::Instant::now();
        if false {
            println!("paint after {:?}",
                     duration_in_ms(now.duration_since(self.last_paint)));
        }
        self.last_paint = now;

        use_color(cr, DEFAULT_BG);

        self.use_font(cr);

        let mut vt = self.vt.lock().unwrap();
        let mut buf = String::with_capacity(80);
        for (row, line) in vt.lines[vt.top..].iter().enumerate() {
            cr.save();
            cr.translate(0.0, (row as f64 * self.font_metrics.height));
            let mut attr = Default::default();
            let mut x = 0.0;
            for (col, cell) in line.iter().enumerate() {
                if cell.attr != attr {
                    self.draw_span(cr, x, attr, &buf);
                    x = col as f64 * self.font_metrics.max_x_advance;
                    attr = cell.attr.clone();
                    buf.clear();
                }
                buf.push(cell.ch);
            }
            if buf.len() > 0 {
                self.draw_span(cr, x, attr, &buf);
                buf.clear();
            }
            cr.restore();
        }

        if !vt.hide_cursor {
            cr.save();
            cr.translate(0.0, ((vt.row - vt.top) as f64 * self.font_metrics.height));
            let (ch, mut attr) = {
                let cell = vt.ensure_pos();
                (cell.ch, cell.attr)
            };
            let inv = attr.inverse();
            attr.set_inverse(!inv);
            let bytes = [ch as u8];
            let span = str::from_utf8(&bytes).unwrap();
            self.draw_span(cr,
                           (vt.col as f64 * self.font_metrics.max_x_advance),
                           attr.clone(),
                           span);
            cr.restore();
        }
    }

    fn key(&mut self, ev: &gdk::EventKey) {
        let buf = translate_key(&ev);
        if self.stdin.send(buf).is_err() {
            println!("can't send");
        }
    }
}

fn translate_key(ev: &gdk::EventKey) -> Box<[u8]> {
    if view::is_modifier_key_event(ev) {
        return Box::new([]);
    }
    let keyval = ev.get_keyval();

    match ev.get_state() {
        gdk::enums::modifier_type::ControlMask => {
            if keyval < 128 {
                match keyval as u8 as char {
                    c if c >= 'a' && c <= 'z' => {
                        return Box::new([(c as u8) - ('a' as u8) + 1]);
                    }
                    '[' => return Box::new([27]),
                    _ => {}
                }
            }
        }
        gdk::enums::modifier_type::Mod1Mask => {
            match gdk::keyval_to_unicode(keyval) {
                Some(u) if u < 128 as char => return Box::new([27, u as u8]),
                _ => {}
            }
        }
        s if s == gdk::ModifierType::empty() || s == gdk::enums::modifier_type::ShiftMask => {
            match gdk::keyval_to_unicode(keyval) {
                Some(u) if u < 128 as char => return Box::new([u as u8]),
                _ => {}
            }
        }
        _ => {}
    }
    println!("unhandled key {:?} {:?} {:?}",
             ev.get_keyval(),
             gdk::keyval_name(ev.get_keyval()),
             ev.get_state());
    return Box::new(['?' as u8]);
}
