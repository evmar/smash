
pub struct ReadLine {
    buf: String,
    ofs: usize,
}

impl ReadLine {
    fn new() -> ReadLine {
        ReadLine {
            buf: String::new(),
            ofs: 0,
        }
    }

    fn insert(&mut self, uni: u32) {
        self.buf.insert(self.ofs, uni as u8 as char);
        self.ofs += 1;
    }

    fn get(&self) -> String {
        self.buf.clone()
    }

    // fn key(&mut self, ev: &gdk::EventKey) {
    //     let uni = gdk::keyval_to_unicode(ev.get_keyval());
    //     match ev.get_state() {
    //         s if s == gdk::ModifierType::empty() ||
    // s == gdk::enums::modifier_type::ShiftMask => {
    //             self.buf.insert(self.ofs, uni as char);
    //         }
    //     }
    // }
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
