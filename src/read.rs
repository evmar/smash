use std::io;
use std::io::Read;
use std::fs;



struct ByteReader<R> {
    r: R,
    buf: Option<u8>,
    err: Option<io::Error>,
}

impl<R: Read> ByteReader<R> {
    fn back(&mut self, b: u8) {
        if self.buf != None {
            panic!("can't back twice");
        }
        self.buf = Some(b);
    }
}

impl<R: Read> Iterator for ByteReader<R> {
    type Item = u8;

    fn next(&mut self) -> Option<u8> {
        match self.buf {
            Some(b) => {
                self.buf = None;
                return Some(b);
            }
            None => {}
        }

        let mut buf = [0];
        match self.r.read(&mut buf) {
            Ok(0) => None,
            Ok(..) => {
                println!("buf {}", buf[0]);
                Some(buf[0])
            }
            Err(e) => {
                self.err = Some(e);
                None
            }
        }
    }
}


struct VTReader<'a, R: 'a + Read> {
    r: &'a mut ByteReader<R>,
}

impl<'a, R: Read> VTReader<'a, R> {
    fn read_escape(&mut self) {
        match self.r.next().unwrap() as char {
            '(' => {
                match self.r.next().unwrap() as char {
                    'B' => {
                        // US ASCII; ignore.
                    }
                    c => {}
                }
            }
            '=' => {
                println!("esc=");
            }
            '>' => {
                println!("esc=");
            }
            '[' => {
                println!("esc[");
            }
            c => panic!("bad escape {}", c),
        }

    }

    fn read(&mut self) {
        match self.r.next() {
            None => {}
            Some(0x1b) => {
                self.read_escape();
            }
            Some(c) => println!("byte {}", c as char),
        }
        match self.r.err {
            Some(ref err) => panic!("err {}", err),
            None => {}
        }
    }
}

fn parse() -> io::Result<()> {
    let f = try!(fs::File::open("top"));
    let mut r = ByteReader {
        r: f,
        buf: None,
        err: None,
    };
    let mut vt = VTReader { r: &mut r };
    loop {
        vt.read();
    }
    Ok(())
}
