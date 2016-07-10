extern crate cairo;
extern crate gdk;

use view;
use view::{Layout, View};
use std::collections::HashMap;
use std::rc::Rc;
use std::cell::RefCell;

trait Delegate {
    fn accept(&mut self, command: &str);
}

pub struct ReadLine {
    buf: String,
    ofs: usize,
    commands: HashMap<String, fn(&mut ReadLine)>,
    bindings: HashMap<String, String>,
    delegate: Option<Box<Delegate>>,
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
    cmds!(
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

// history
"accept-line" => rl {
    if let Some(ref mut delegate) = rl.delegate {
        delegate.accept("foo");
    }
}

// changing text
"backward-delete-char" => rl {
    if rl.ofs > 0 {
        rl.buf.remove(rl.ofs - 1);
        rl.ofs -= 1;
    }
})
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
        "C-b" => "backward-char",

        "Return" => "accept-line",

        "BackSpace" => "backward-delete-char"
    )
}

impl ReadLine {
    pub fn new() -> ReadLine {
        ReadLine {
            buf: String::new(),
            ofs: 0,
            commands: make_command_map(),
            bindings: make_binds_map(),
            delegate: None,
        }
    }

    pub fn insert(&mut self, text: &str) {
        for c in text.as_bytes() {
            self.buf.insert(self.ofs, *c as char);
            self.ofs += 1;
        }
    }

    pub fn key(&mut self, key: &str) -> bool {
        // This function has an odd control flow because we need to run
        // the keybinding command without any of the hashtables holding
        // an immutable borrow on self.
        let f = {
            match self.bindings.get(key) {
                None => None,
                Some(command) => {
                    match self.commands.get(command) {
                        None => {
                            println!("no command named {:?}", command);
                            return false;
                        }
                        Some(f) => Some(*f),
                    }
                }
            }
        };
        if let Some(f) = f {
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
}

pub struct ReadLineView {
    context: view::ContextRef,
    pub rl: ReadLine,
}

impl ReadLineView {
    pub fn new(context: view::ContextRef) -> Rc<RefCell<ReadLineView>> {
        let view = Rc::new(RefCell::new(ReadLineView {
            context: context,
            rl: ReadLine::new(),
        }));
        let delegate = Box::new(view.clone());
        view.borrow_mut().rl.delegate = Some(delegate);
        view
    }

    fn use_font(&mut self, cr: &cairo::Context) {
        cr.select_font_face("sans",
                            cairo::enums::FontSlant::Normal,
                            cairo::enums::FontWeight::Normal);
        cr.set_font_size(18.0);
    }
}

impl View for ReadLineView {
    fn draw(&mut self, cr: &cairo::Context) {
        self.use_font(cr);
        cr.set_source_rgb(0.0, 0.0, 0.0);
        let ext = cr.font_extents();

        cr.translate(0.0, ext.ascent);
        let str = self.rl.buf.as_str();
        cr.show_text(str);

        let text_ext = cr.text_extents(&str[0..self.rl.ofs]);
        cr.rectangle(text_ext.x_advance,
                     -ext.ascent,
                     3.0,
                     ext.ascent + ext.descent);
        cr.fill();
    }

    fn key(&mut self, ev: &gdk::EventKey) {
        if let Some(key) = translate_key(ev) {
            if self.rl.key(&key) {
                self.context.borrow_mut().dirty();
            }
        }
    }

    fn layout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        self.use_font(cr);
        let ext = cr.font_extents();
        Layout {
            width: space.width,
            height: ext.height.round() as i32,
        }
    }
}

impl Delegate for Rc<RefCell<ReadLineView>> {
    fn accept(&mut self, _command: &str) {
        let rl = self.clone();
        view::add_task(Box::new(move || {
            let mut rl = rl.borrow_mut();
            rl.rl.clear();
            rl.context.borrow_mut().dirty();
        }));
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
}
