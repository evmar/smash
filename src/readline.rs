extern crate cairo;
extern crate gdk;

use view;
use view::View;

pub struct ReadLine {
    buf: String,
    ofs: usize,
}

impl ReadLine {
    pub fn new() -> ReadLine {
        ReadLine {
            buf: String::new(),
            ofs: 0,
        }
    }

    pub fn insert(&mut self, uni: char) {
        self.buf.insert(self.ofs, uni);
        self.ofs += 1;
    }

    pub fn backspace(&mut self) {
        if self.ofs > 0 {
            self.buf.remove(self.ofs - 1);
            self.ofs -= 1;
        }
    }

    pub fn get(&self) -> String {
        self.buf.clone()
    }
}

pub struct ReadLineView {
    context: view::ContextRef,
    pub rl: ReadLine,
}

impl ReadLineView {
    pub fn new(context: view::ContextRef) -> ReadLineView {
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
        cr.rectangle(text_ext.width + 2.0,
                     -ext.ascent,
                     3.0,
                     ext.ascent + ext.descent);
        cr.fill();
    }

    fn key(&mut self, ev: &gdk::EventKey) {
        if view::is_modifier_key_event(ev) {
            return;
        }
        match ev.get_state() {
            s if s == gdk::ModifierType::empty() || s == gdk::enums::modifier_type::ShiftMask => {
                if let Some(uni) = gdk::keyval_to_unicode(ev.get_keyval()) {
                    match uni {
                        '\x08' => {
                            self.rl.backspace();
                            self.context.borrow_mut().dirty = true;
                            return;
                        }
                        uni if uni >= ' ' => {
                            self.rl.insert(uni);
                            self.context.borrow_mut().dirty = true;
                            return;
                        }
                        _ => {
                            println!("bad uni: {:?}", uni);
                        }
                    }
                }
            }
            _ => {}
        }
        let key = ev.as_ref();
        println!("unhandled key: state:{:?} val:{:?}",
                 key.state,
                 gdk::keyval_name(key.keyval));
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn append() {
        let mut rl = ReadLine::new();
        rl.insert('a');
        assert_eq!("a", rl.get());
        rl.insert('b');
        assert_eq!("ab", rl.get());
    }
}
