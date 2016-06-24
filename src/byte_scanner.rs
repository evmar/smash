use std::io;
use std::io::Read;

/// ByteScanner supports reading a byte at a time from an io::Read
/// as well as backing up by one byte.  Internally it buffers.
pub struct ByteScanner {
    buf: [u8; 4 << 10],
    ofs: usize,
    len: usize,
}

impl ByteScanner {
    pub fn new() -> ByteScanner {
        ByteScanner {
            buf: [0; 4 << 10],
            ofs: 0,
            len: 0,
        }
    }

    pub fn read<R: Read>(&mut self, r: &mut R) -> io::Result<bool> {
        if self.ofs < self.len {
            for i in 0..(self.len - self.ofs) {
                self.buf[i] = self.buf[self.ofs + i];
            }
            self.ofs = self.len - self.ofs;
        } else {
            self.ofs = 0;
        }
        self.len = self.ofs + try!(r.read(&mut self.buf[self.ofs..]));
        self.ofs = 0;
        return Ok(self.len > self.ofs);
    }

    pub fn back(&mut self) {
        if self.ofs == 0 {
            panic!("can't back")
        }
        self.ofs -= 1;
    }

    pub fn mark(&mut self) -> usize {
        self.ofs
    }
    pub fn pop_mark(&mut self, n: usize) {
        self.ofs = n;
    }

    pub fn next(&mut self) -> Option<u8> {
        if self.ofs == self.len {
            return None;
        }
        let c = self.buf[self.ofs];
        self.ofs += 1;
        return Some(c);
    }
}

// Non-consuming iterator implementation.
// impl<'a> Iterator for &'a mut ByteScanner {
//     type Item = u8;
//     fn next(&mut self) -> Option<u8> {
//         (*self).next()
//     }
// }
