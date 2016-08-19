extern crate libc;

use std::io::Read;
use std::collections::hash_set::HashSet;
use std::sync::Mutex;
use byte_scanner::ByteScanner;

const EIO: libc::c_int = 5;

macro_rules! probe {
    ( $x:expr ) => {{
        match $x {
            Some(x) => x,
            None => return None
        }
    }}
}

fn bprefix(b: u8, n: usize, prefix: u8) -> Option<u8> {
    if b >> 8 - n == prefix {
        Some(b & ((1 << (8 - n)) - 1))
    } else {
        None
    }
}

// xxxx xxIB AAAA CCCC
//  I = inverse
//  B = bright
//  A = background color
//  C = foreground color
#[derive(Clone)]
#[derive(Copy)]
#[derive(Debug)]
#[derive(Default)]
#[derive(PartialEq)]
pub struct Attr {
    pub val: u16,
}

impl Attr {
    pub fn inverse(&self) -> bool {
        self.val >> 9 & 1 != 0
    }
    pub fn set_inverse(&mut self, set: bool) {
        if set {
            self.val |= 1 << 9
        } else {
            self.val &= !(1 << 9)
        }
    }
    pub fn bold(&self) -> bool {
        self.val >> 8 & 1 != 0
    }
    pub fn set_bold(&mut self) {
        self.val |= 1 << 8
    }

    pub fn fg(&self) -> Option<usize> {
        match self.val & 0x000F {
            0 => None,
            v => Some(v as usize - 1),
        }
    }
    pub fn set_fg(&mut self, color: Option<usize>) {
        let val = match color {
            None => 0,
            Some(c) if c < 8 => c + 1,
            Some(c) => panic!("bad color {}", c),
        };
        self.val = self.val & 0xFFF0 | val as u16;
    }
    pub fn bg(&self) -> Option<usize> {
        match self.val >> 4 & 0x000F {
            0 => None,
            v => Some(v as usize - 1),
        }
    }
    pub fn set_bg(&mut self, color: Option<usize>) {
        let val = match color {
            None => 0,
            Some(c) if c < 8 => c + 1,
            Some(c) => panic!("bad color {}", c),
        };
        self.val = self.val & 0xFF0F | ((val as u16) << 4);
    }
}

#[derive(Debug)]
#[derive(Copy)]
#[derive(Clone)]
pub struct Cell {
    pub ch: char,
    pub attr: Attr,
}

pub struct VT {
    pub lines: Vec<Box<Vec<Cell>>>,
    /// The number of lines visible on screen.
    pub height: usize,
    /// The number of columns visible on screen.
    pub width: usize,
    /// The index of the first visible line on screen.
    pub top: usize,
    /// The index of the line the cursor is on, relative to the lines buffer.
    pub row: usize,
    /// The column the cursor is on.
    pub col: usize,
    /// True if the cursor shouldn't be displayed.
    pub hide_cursor: bool,
    /// The current display attributes, to be used when text is added.
    pub attr: Attr,
}

impl VT {
    pub fn new() -> VT {
        VT {
            lines: Vec::new(),
            height: 25,
            width: 80,
            top: 0,
            row: 0,
            col: 0,
            hide_cursor: false,
            attr: Attr { val: 0 },
        }
    }

    pub fn ensure_pos(&mut self) -> &mut Cell {
        if self.col >= self.width {
            self.row += 1;
            self.col = 0;
        }
        while self.row >= self.lines.len() {
            self.lines.push(Box::new(Vec::with_capacity(self.width)));
        }
        let mut row = &mut self.lines[self.row];
        while self.col >= row.len() {
            row.push(Cell {
                ch: ' ',
                attr: Default::default(),
            });
        }
        if self.row >= self.top + self.height {
            self.top = self.row - self.height + 1;
        }
        return &mut row[self.col];
    }

    // Drop the last line if it's empty; used when a subprocess exits.
    pub fn trim(&mut self) {
        if self.col == 0 && self.row > 0 && self.row >= self.lines.len() {
            self.row -= 1;
            if self.row < self.lines.len() - 1 {
                self.lines.pop();
            }
        }
    }

    #[allow(dead_code)]
    pub fn dump(&self) -> String {
        let mut buf = String::new();
        println!("dump len {}", self.lines.len());
        for line in self.lines.iter() {
            for cell in line.iter() {
                buf.push(cell.ch)
            }
            buf.push('\n')
        }
        buf
    }
}

pub struct VTReader<'a> {
    todo: HashSet<String>,
    vt: &'a Mutex<VT>,
    r: ByteScanner,
    stdin: Box<FnMut(Box<[u8]>)>,
}

impl<'a> VTReader<'a> {
    pub fn new(vt: &'a Mutex<VT>, stdin: Box<FnMut(Box<[u8]>)>) -> VTReader<'a> {
        VTReader {
            todo: HashSet::new(),
            vt: vt,
            r: ByteScanner::new(),
            stdin: stdin,
        }
    }

    pub fn read<R: Read>(&mut self, r: &mut R) -> bool {
        if let Err(ref err) = self.r.read(r) {
            if err.raw_os_error().unwrap() == EIO {
                return false;
            }
            panic!("read failed: {:?}", err);
        }

        let mut vt = self.vt.lock().unwrap();
        let mut mark;
        let todo = &mut self.todo;
        let todo = &mut |msg: String| {
            if todo.insert(msg.clone()) {
                println!("TODO: {}", msg)
            }
        };
        loop {
            mark = self.r.mark();
            let mut vtr = VTRead {
                todo: todo,
                vt: &mut vt,
                r: &mut self.r,
                stdin: &mut *self.stdin,
            };
            match vtr.read() {
                None => {
                    break;
                }
                Some(_) => {}
            }
        }
        self.r.pop_mark(mark);
        return true;
    }
}

pub struct VTRead<'a> {
    todo: &'a mut FnMut(String),
    vt: &'a mut VT,
    r: &'a mut ByteScanner,
    stdin: &'a mut FnMut(Box<[u8]>),
}

impl<'a> VTRead<'a> {
    fn todo<S: Into<String>>(&mut self, msg: S) {
        (self.todo)(msg.into());
    }
    fn read_num(&mut self) -> Option<usize> {
        let mut num = 0;
        loop {
            let c = probe!(self.r.next());
            match c as char {
                '0'...'9' => num = num * 10 + ((c - ('0' as u8)) as usize),
                _ => {
                    self.r.back();
                    return Some(num);
                }
            }
        }
    }

    fn read_csi(&mut self) -> Option<()> {
        let mut question = false;
        let mut gt = false;

        let mut args: [usize; 2] = [0; 2];
        let mut argc = 0;
        loop {
            match probe!(self.r.next()) as char {
                '?' => question = true,
                '>' => gt = true,
                '0'...'9' => {
                    self.r.back();
                    if argc == args.len() {
                        panic!("too many args")
                    }
                    args[argc] = probe!(self.read_num());
                    argc += 1;
                }
                ';' => {}
                _ => {
                    self.r.back();
                    break;
                }
            }
        }
        let args = &args[..argc];

        let cmd = probe!(self.r.next()) as char;
        match cmd {
            '@' => {
                // insert blanks
                let count = *args.get(0).unwrap_or(&1);
                for _ in 0..count {
                    self.vt.ensure_pos();
                    self.vt.col += 1;
                }
            }
            'A' => {
                // cursor up
                let dy = *args.get(0).unwrap_or(&1);
                if self.vt.row > dy {
                    self.vt.row -= dy;
                }
            }
            'B' => {
                // cursor down
                let dy = *args.get(0).unwrap_or(&1);
                self.vt.row += dy;
            }
            'C' => {
                // cursor forward
                let dx = *args.get(0).unwrap_or(&1);
                self.vt.col += dx;
            }
            'D' => {
                // cursor back
                let dx = *args.get(0).unwrap_or(&1);
                if self.vt.col > dx {
                    self.vt.col -= dx;
                }
            }
            'H' => {
                let row = *args.get(0).unwrap_or(&1) - 1;
                self.vt.row = self.vt.top + row;
                self.vt.col = *args.get(1).unwrap_or(&1) - 1;
            }
            'J' => {
                match *args.get(0).unwrap_or(&0) {
                    2 => {
                        let (top, height) = (self.vt.top, self.vt.height);
                        let end = if top + height > self.vt.lines.len() {
                            self.vt.lines.len()
                        } else {
                            top + height
                        };
                        for line in self.vt.lines[top..end].iter_mut() {
                            line.clear();
                        }
                    }
                    x => self.todo(format!("erase in display {}", x)),
                }
            }
            'K' => {
                match *args.get(0).unwrap_or(&0) {
                    0 => {
                        // Erase to Right.
                        if self.vt.row < self.vt.lines.len() {
                            let row = self.vt.row;
                            if self.vt.col < self.vt.lines[row].len() {
                                let col = self.vt.col;
                                self.vt.lines[row].truncate(col);
                            }
                        }
                    }
                    mode => {
                        self.todo(format!("erase in line {:?}", mode));
                    }
                }
            }
            'L' => {
                self.vt.ensure_pos();
                let row = self.vt.row;
                self.vt.lines.insert(row, Box::new(Vec::new()));
            }
            'P' => {
                let count = *args.get(0).unwrap_or(&1);
                let (row, col) = (self.vt.row, self.vt.col);
                let line = &mut self.vt.lines[row];
                for i in 0..count {
                    if col + count + i >= line.len() {
                        break;
                    }
                    line[col + i] = line[col + count + i];
                }
                let newlen = line.len() - count;
                line.truncate(newlen - count);
            }
            'c' if gt => {
                // send device attributes (secondary)
                (self.stdin)(b"\x1b[41;0;0c".to_vec().into_boxed_slice());
            }
            'd' => {
                let row = *args.get(0).unwrap_or(&1) - 1;
                self.vt.row = self.vt.top + row;
            }
            'h' | 'l' if question => {
                let set = cmd == 'h';
                match args[0] {
                    1 => self.todo("application cursor keys mode"),
                    12 => self.todo("start blinking cursor"),
                    25 => self.vt.hide_cursor = !set,
                    1049 => {
                        self.todo("save cursor");
                        self.todo("alternate screen buffer");
                    }
                    arg => self.todo(format!("?h arg {}", arg)),
                }
            }
            'h' | 'l' => self.todo(format!("re/set mode {}", args[0])),
            'm' => {
                // character attributes
                if args.len() == 0 {
                    self.vt.attr.val = 0;
                }
                for attr in args {
                    match *attr {
                        0 => self.vt.attr.val = 0,
                        1 => self.vt.attr.set_bold(),
                        7 => self.vt.attr.set_inverse(true),
                        27 => self.vt.attr.set_inverse(false),
                        v if v >= 30 && v < 39 => self.vt.attr.set_fg(Some(v - 30)),
                        39 => self.vt.attr.set_fg(None),
                        v if v >= 40 && v < 49 => self.vt.attr.set_bg(Some(v - 40)),
                        49 => self.vt.attr.set_bg(None),
                        _ => {
                            self.todo(format!("set attr {}", attr));
                        }
                    }
                }
            }
            'n' => {
                // device status report
                if args.len() == 1 {
                    match args[0] {
                        5 => {
                            // status report
                            self.todo("status report");
                        }
                        6 => {
                            // report cursor position
                            self.todo("cursor pos");
                        }
                        _ => {
                            self.todo(format!("device status {}", args[0]));
                        }
                    }
                }
            }
            'r' => {
                // set scrolling region
                let top = *args.get(0).unwrap_or(&1);
                let bottom = *args.get(1).unwrap_or(&self.vt.height);
                if top == 1 && bottom == self.vt.height {
                    // full window
                } else {
                    self.todo(format!("set scrolling region: {}:{}", top, bottom));
                }
            }
            c => self.todo(format!("unhandled CSI {}", c)),
        }
        return Some(());
    }

    fn read_osc(&mut self) -> Option<()> {
        let cmd = probe!(self.read_num());
        match probe!(self.r.next()) as char {
            ';' => {}
            _ => panic!("unexpected OSC"),
        }
        let mut text = String::new();
        loop {
            match probe!(self.r.next()) {
                0 | 7 => break,
                c => text.push(c as char),
            }
        }
        match cmd {
            0 => self.todo(format!("todo: set title+icon to {:?}", text)),
            1 => self.todo(format!("todo: set icon to {:?}", text)),
            2 => self.todo(format!("todo: set title to {:?}", text)),
            11 => {
                self.todo(format!("todo: vt100 background control {:?}", text));
            }
            _ => self.todo(format!("todo: osc {} {:?}", cmd, text)),
        }
        Some(())
    }

    fn read_escape(&mut self) -> Option<()> {
        match probe!(self.r.next()) as char {
            '(' => {
                match probe!(self.r.next()) as char {
                    'B' => {
                        // US ASCII
                    }
                    c => self.todo(format!("g0 charset {}", c)),
                }
            }
            '=' => self.todo("application keypad"),
            '[' => probe!(self.read_csi()),
            ']' => probe!(self.read_osc()),
            '>' => self.todo("normal keypad"),
            'M' => {
                // move up/insert line
                self.vt.ensure_pos();
                let top = self.vt.top;
                self.vt.lines.insert(top, Box::new(Vec::new()));
            }
            c => panic!("notimpl: esc {}", c),
        }
        return Some(());
    }

    fn read_utf8(&mut self) -> Option<u32> {
        let c = probe!(self.r.next());
        let (n, mut rune) = {
            if let Some(c) = bprefix(c, 3, 0b110) {
                (1, c as u32)
            } else if let Some(c) = bprefix(c, 4, 0b1110) {
                (2, c as u32)
            } else {
                panic!("read_utf8 {}", c);
            }
        };

        for _ in 0..n {
            let c = probe!(self.r.next());
            if let Some(c) = bprefix(c, 2, 0b10) {
                rune = rune << 6 | c as u32;
            } else {
                self.r.back();
                panic!("read_utf8 continuation {}", c);
            }
        }
        Some(rune)
    }

    fn read(&mut self) -> Option<()> {
        match probe!(self.r.next()) as char {
            '\x07' => self.todo("bell"),
            '\x08' => {
                if self.vt.col > 0 {
                    self.vt.col -= 1;
                }
            }
            '\x09' => {
                self.vt.col += 8 - (self.vt.col % 8);
            }
            '\n' => {
                self.vt.row += 1;
                self.vt.col = 0;
            }
            '\r' => {
                self.vt.col = 0;
            }
            '\x1b' => probe!(self.read_escape()),
            c if c >= ' ' => {
                let ch = {
                    if c as u8 >= 0x80 {
                        self.r.back();
                        let _ = probe!(self.read_utf8());
                        '?'
                    } else {
                        c
                    }
                };

                *self.vt.ensure_pos() = Cell {
                    ch: ch,
                    attr: self.vt.attr.clone(),
                };
                self.vt.col += 1;
            }
            c => {
                panic!("unhandled input {:?}", c);
            }
        }
        Some(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::Mutex;

    fn write_text(vt: &Mutex<VT>, text: &[u8]) {
        let mut r = VTReader::new(&vt, Box::new(|_| {}));
        r.read(&mut &text[..]);
    }

    #[test]
    fn utf8() {
        let vt = Mutex::new(VT::new());
        write_text(&vt,
                   &[0xE6, 0x97, 0xA5 /* 日 */, 0xE6, 0x9C, 0xAC /* 本 */, 0xE8,
                     0xAA, 0x9E /* 語 */]);
        assert_eq!(vt.lock().unwrap().col, 3);
    }

    #[test]
    fn trim() {
        let vt = Mutex::new(VT::new());
        write_text(&vt, "hello, world\n".as_bytes());
        vt.lock().unwrap().trim();
        let vt = vt.lock().unwrap();
        vt.dump();
        assert_eq!(vt.row, 0);
        assert_eq!(vt.col, 0);
        assert_eq!(vt.top, 0);
        assert_eq!(vt.lines.len(), 1);
    }
}
