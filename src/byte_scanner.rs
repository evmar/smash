extern crate libc;

use std::io::Read;

const EIO: libc::c_int = 5;

/// ByteScanner supports reading a byte at a time from an io::Read
/// as well as backing up by one byte.  Internally it buffers.
pub struct ByteScanner<R: Read> {
    buf: [u8; 4 << 10],
    ofs: usize,
    len: usize,
    r: R,
}

impl<R: Read> ByteScanner<R> {
    pub fn new(r: R) -> ByteScanner<R> {
        ByteScanner {
            buf: [0; 4 << 10],
            ofs: 0,
            len: 0,
            r: r,
        }
    }

    fn fill(&mut self) -> bool {
        self.ofs = 0;
        self.len = match self.r.read(&mut self.buf) {
            Err(ref err) if err.raw_os_error().unwrap() == EIO => 0,
            Err(err) => {
                panic!("read failed: {:?}", err);
            }
            Ok(len) => len,
        };
        return self.len > self.ofs;
    }

    pub fn back(&mut self) {
        if self.ofs == 0 {
            panic!("can't back")
        }
        self.ofs -= 1;
    }

    pub fn next(&mut self) -> Option<u8> {
        if self.ofs == self.len {
            if !self.fill() {
                return None;
            }
        }
        let c = self.buf[self.ofs];
        self.ofs += 1;
        return Some(c);
    }
}

/// Non-consuming iterator implementation.
impl<'a, R: Read> Iterator for &'a mut ByteScanner<R> {
    type Item = u8;
    fn next(&mut self) -> Option<u8> {
        (*self).next()
    }
}
