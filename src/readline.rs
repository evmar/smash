extern crate cairo;
extern crate gdk;

use view;
use view::{Layout, View};
use std::collections::HashMap;
use std::rc::Rc;
use std::cell::RefCell;

struct Config {
    commands: HashMap<String, fn(&mut ReadLine)>,
    bindings: HashMap<String, String>,
}

pub struct ReadLine {
    config: Rc<Config>,
    buf: String,
    ofs: usize,
    pub accept_cb: Box<FnMut(&str)>,
}

macro_rules! cmds {
    ($($name:expr => $var:ident $body:expr)*) => {{
        let mut map: HashMap<String, fn(&mut ReadLine)> = HashMap::new();
        $(
            {
                fn f($var: &mut ReadLine) {$body}
                map.insert(String::from($name), f);
            }
        )*;
        map
    }}
}

fn make_command_map() -> HashMap<String, fn(&mut ReadLine)> {
    return cmds!(
    // movement
"beginning-of-line" => rl {
    rl.ofs = 0;
}
"end-of-line" => rl {
    rl.ofs = rl.buf.len();
}
"forward-char" => rl {
    if rl.ofs < rl.buf.len() {
        rl.ofs += 1;
    }
}
"backward-char" => rl {
    if rl.ofs > 0 {
        rl.ofs -= 1;
    }
}
"forward-word" => rl {
    while rl.ofs < rl.buf.len() && !rl.in_word() {
        rl.ofs += 1;
    }
    while rl.ofs < rl.buf.len() && rl.in_word() {
        rl.ofs += 1;
    }
}
"backward-word" => rl {
    if rl.ofs > 0 {
        rl.ofs -= 1;
    }
    while rl.ofs > 0 && !rl.in_word() {
        rl.ofs -= 1;
    }
    while rl.ofs > 0 && rl.in_word() {
        rl.ofs -= 1;
    }
    if rl.ofs < rl.buf.len() && !rl.in_word() {
        rl.ofs += 1;
    }
}

    // history
"accept-line" => rl {
    (rl.accept_cb)(&rl.buf);
}

    // changing text
"backward-delete-char" => rl {
    if rl.ofs > 0 {
        rl.buf.remove(rl.ofs - 1);
        rl.ofs -= 1;
    }
}

    // killing and yanking
"kill-line" => rl {
    rl.buf.truncate(rl.ofs);
}
"unix-line-discard" => rl {
    rl.buf.truncate(0);
    rl.ofs = 0;
}
"backward-kill-word" => rl {
    let end = rl.ofs;
    rl.config.command("backward-word").unwrap()(rl);
    let start = rl.ofs;
    let mut new_buf = String::new();
    for i in 0..start {
        new_buf.push(rl.buf.as_bytes()[i] as char);
    }
    for i in end..rl.buf.len() {
        new_buf.push(rl.buf.as_bytes()[i] as char);
    }
    rl.buf = new_buf;
});
}


macro_rules! binds {
    ( $( $key:expr => $cmd:expr ),* ) => {{
        let mut map: HashMap<String, String> = HashMap::new();
        $(
            map.insert(String::from($key),
                       String::from($cmd));
        )*;
        map
    }}
}

fn make_binds_map() -> HashMap<String, String> {
    binds!(
        "C-a" => "beginning-of-line",
        "C-e" => "end-of-line",
        "C-f" => "forward-char",
        "Right" => "forward-char",
        "C-b" => "backward-char",
        "Left" => "backward-char",
        "M-f" => "forward-word",
        "M-b" => "backward-word",

        "Return" => "accept-line",

        "BackSpace" => "backward-delete-char",

        "C-k" => "kill-line",
        "C-u" => "unix-line-discard",
        "M-BackSpace" => "backward-kill-word"
    )
}

impl Config {
    fn key(&self, key: &str) -> Option<fn(&mut ReadLine)> {
        self.bindings.get(key).and_then(|name| self.command(name))
    }
    fn command(&self, name: &str) -> Option<fn(&mut ReadLine)> {
        self.commands.get(name).map(|f| *f)
    }
}

impl ReadLine {
    pub fn new() -> ReadLine {
        let config = Rc::new(Config {
            commands: make_command_map(),
            bindings: make_binds_map(),
        });
        ReadLine {
            config: config,
            buf: String::new(),
            ofs: 0,
            accept_cb: Box::new(move |_: &str| {}),
        }
    }

    pub fn insert(&mut self, text: &str) {
        for c in text.as_bytes() {
            self.buf.insert(self.ofs, *c as char);
            self.ofs += 1;
        }
    }

    pub fn key(&mut self, key: &str) -> bool {
        if let Some(f) = self.config.key(key) {
            f(self);
            return true;
        }

        if key.len() == 1 {
            self.insert(key);
            return true;
        }

        println!("no binding for {:?}", key);
        return false;
    }

    pub fn get(&self) -> String {
        self.buf.clone()
    }

    pub fn clear(&mut self) {
        self.buf.clear();
        self.ofs = 0;
    }

    fn in_word(&self) -> bool {
        if self.ofs >= self.buf.len() {
            return false;
        }
        let c = self.buf.as_bytes()[self.ofs] as char;
        return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9');
    }
}

pub struct ReadLineView {
    pub dirty: Rc<Fn()>,
    pub rl: ReadLine,
    layout: Layout,
}

impl ReadLineView {
    pub fn new(dirty: Rc<Fn()>) -> Rc<RefCell<ReadLineView>> {
        Rc::new(RefCell::new(ReadLineView {
            dirty: dirty,
            rl: ReadLine::new(),
            layout: Layout::new(),
        }))
    }

    fn use_font(&self, cr: &cairo::Context) {
        cr.select_font_face("sans",
                            cairo::enums::FontSlant::Normal,
                            cairo::enums::FontWeight::Normal);
        cr.set_font_size(18.0);
    }
}

impl View for ReadLineView {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        let rl = self;
        rl.use_font(cr);
        cr.set_source_rgb(0.0, 0.0, 0.0);
        let ext = cr.font_extents();

        cr.translate(0.0, ext.ascent);
        let rl = &rl.rl;
        let str = rl.buf.as_str();
        cr.show_text(str);

        if focus {
            let text_ext = cr.text_extents(&str[0..rl.ofs]);
            cr.rectangle(text_ext.x_advance,
                         -ext.ascent,
                         2.0,
                         ext.ascent + ext.descent);
            cr.fill();
        }
    }

    fn key(&mut self, ev: &gdk::EventKey) {
        let rl = self;
        if let Some(key) = translate_key(ev) {
            if rl.rl.key(&key) {
                (rl.dirty)();
            }
        }
    }

    fn relayout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        self.use_font(cr);
        let ext = cr.font_extents();
        self.layout = Layout {
            width: space.width,
            height: ext.height.round() as i32,
        };
        self.layout
    }
    fn get_layout(&self) -> Layout {
        self.layout
    }
}

fn translate_key(ev: &gdk::EventKey) -> Option<String> {
    if view::is_modifier_key_event(ev) {
        return None;
    }

    let mut name = String::new();
    if ev.get_state().contains(gdk::enums::modifier_type::ControlMask) {
        name.push_str("C-");
    }
    if ev.get_state().contains(gdk::enums::modifier_type::Mod1Mask) {
        name.push_str("M-");
    }

    if name.len() == 0 {
        if let Some(uni) = gdk::keyval_to_unicode(ev.get_keyval()) {
            if uni >= ' ' {
                return Some(uni.to_string());
            }
        }
    }

    let gdkname = match gdk::keyval_name(ev.get_keyval()) {
        Some(n) => n,
        None => {
            println!("unnamed key {}", ev.get_keyval());
            return None;
        }
    };
    name.push_str(&gdkname);

    // name.push_str(&gdk::keyval_name(ev.get_keyval()).unwrap());

    if name.len() == 0 {
        let key = ev.as_ref();
        println!("unhandled key: state:{:?} val:{:?}",
                 key.state,
                 gdk::keyval_name(key.keyval));
        return None;
    }

    return Some(name);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn append() {
        let mut rl = ReadLine::new();
        rl.insert("a");
        assert_eq!("a", rl.get());
        rl.insert("bc");
        assert_eq!("abc", rl.get());
    }

    #[test]
    fn forward_word() {
        let mut rl = ReadLine::new();
        rl.key("M-f");
        assert_eq!(rl.ofs, 0);

        rl.insert("012 456  9");
        rl.ofs = 3;
        rl.key("M-f");
        assert_eq!(rl.ofs, 7);
        rl.key("M-f");
        assert_eq!(rl.ofs, 10);
        rl.key("M-f");
        assert_eq!(rl.ofs, 10);
    }

    #[test]
    fn backward_word() {
        let mut rl = ReadLine::new();
        rl.key("M-b");
        assert_eq!(rl.ofs, 0);

        rl.insert("012 456  9");
        rl.ofs = 10;
        rl.key("M-b");
        assert_eq!(rl.ofs, 9);
        rl.key("M-b");
        assert_eq!(rl.ofs, 4);
        rl.key("M-b");
        assert_eq!(rl.ofs, 0);
        rl.key("M-b");
        assert_eq!(rl.ofs, 0);
    }
}
