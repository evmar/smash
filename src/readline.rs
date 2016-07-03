extern crate cairo;
extern crate gdk;

use view;
use view::View;
use std::collections::HashMap;

pub struct ReadLine {
    buf: String,
    ofs: usize,
    commands: HashMap<String, fn(&mut ReadLine)>,
    bindings: HashMap<String, String>,
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

        "Backspace" => "backward-delete-char"
    )
}

#[derive(Debug)]
pub enum Key {
    Special(String),
    Text(String),
}

impl ReadLine {
    pub fn new() -> ReadLine {
        ReadLine {
            buf: String::new(),
            ofs: 0,
            commands: make_command_map(),
            bindings: make_binds_map(),
        }
    }

    pub fn insert(&mut self, text: &str) {
        for c in text.as_bytes() {
            self.buf.insert(self.ofs, *c as char);
            self.ofs += 1;
        }
    }

    pub fn key(&mut self, key: Key) -> bool {
        match key {
            Key::Text(text) => {
                self.insert(&text);
                return true;
            }
            Key::Special(ref name) => {
                let f = {
                    let command = match self.bindings.get(name) {
                        None => {
                            println!("no binding for {:?}", name);
                            return false;
                        }
                        Some(command) => command,
                    };
                    match self.commands.get(command) {
                        None => {
                            println!("no command named {:?}", command);
                            return false;
                        }
                        Some(f) => *f,
                    }
                };
                f(self);
                return true;
            }
        };
    }

    pub fn get(&self) -> String {
        self.buf.clone()
    }
}

pub struct ReadLineView {
    context: Box<view::Context>,
    pub rl: ReadLine,
}

impl ReadLineView {
    pub fn new(context: Box<view::Context>) -> ReadLineView {
        ReadLineView {
            context: context,
            rl: ReadLine::new(),
        }
    }
}

impl View for ReadLineView {
    fn draw(&mut self, cr: &cairo::Context) {
        cr.set_source_rgb(0.0, 0.0, 0.0);
        cr.select_font_face("sans",
                            cairo::enums::FontSlant::Normal,
                            cairo::enums::FontWeight::Normal);
        cr.set_font_size(18.0);
        let ext = cr.font_extents();

        cr.translate(0.0, ext.height);
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
            if self.rl.key(key) {
                self.context.dirty();
            }
        }
    }
}

fn translate_key(ev: &gdk::EventKey) -> Option<Key> {
    if view::is_modifier_key_event(ev) {
        return None;
    }

    let mut special = false;
    let mut name = String::new();
    if ev.get_state().contains(gdk::enums::modifier_type::ControlMask) {
        name.push_str("C-");
        special = true;
    }
    if ev.get_state().contains(gdk::enums::modifier_type::Mod1Mask) {
        name.push_str("M-");
        special = true;
    }

    match gdk::keyval_to_unicode(ev.get_keyval()) {
        Some(uni) => {
            match uni {
                '\x08' => {
                    special = true;
                    name.push_str("Backspace");
                }
                '\x09' => {
                    if special {
                        name.push_str("Tab");
                    } else {
                        name.push_str("\x09");
                    }
                }
                ' ' => {
                    if special {
                        name.push_str("Space");
                    } else {
                        name.push_str(" ");
                    }
                }
                uni if uni > ' ' => {
                    name.push_str(&gdk::keyval_name(ev.get_keyval()).unwrap());
                }
                _ => {
                    println!("bad uni: {:?}", uni);
                }
            }
        }
        None => {}
    }

    if name.len() > 0 {
        return Some(if special {
            Key::Special(name)
        } else {
            Key::Text(name)
        });
    }

    let key = ev.as_ref();
    println!("unhandled key: state:{:?} val:{:?}",
             key.state,
             gdk::keyval_name(key.keyval));
    return None;
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
