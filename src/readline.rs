extern crate cairo;
extern crate gdk;

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

    pub fn get(&self) -> String {
        self.buf.clone()
    }
}

pub struct ReadLineView {
    pub rl: ReadLine,
}
impl ReadLineView {
    pub fn new() -> ReadLineView {
        ReadLineView { rl: ReadLine::new() }
    }
    pub fn draw(&mut self, cr: &cairo::Context) {
        cr.set_source_rgb(0.0, 0.0, 0.0);
        cr.select_font_face("sans",
                            cairo::enums::FontSlant::Normal,
                            cairo::enums::FontWeight::Normal);
        cr.set_font_size(18.0);
        cr.move_to(20.0, 20.0);
        cr.show_text(self.rl.buf.as_str());
    }

    pub fn key(&mut self, ev: &gdk::EventKey) {
        match ev.get_state() {
            s if s == gdk::ModifierType::empty() || s == gdk::enums::modifier_type::ShiftMask => {
                match gdk::keyval_to_unicode(ev.get_keyval()) {
                    Some(c) if c >= ' ' => {
                        self.rl.insert(c);
                    }
                    _ => {}
                }
            }
            _ => {}
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn append() {
        let mut rl = ReadLine::new();
        rl.insert('a' as u32);
        assert_eq!("a", rl.get());
        rl.insert('b' as u32);
        assert_eq!("ab", rl.get());
    }
}
