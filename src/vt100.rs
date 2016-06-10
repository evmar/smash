use std::fmt::Display;
use std::fs;
use std::io::Write;
use std::collections::hash_set::HashSet;
use std::sync::Mutex;
use byte_scanner::ByteScanner;

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
            Some(c) => c + 1,
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
            Some(c) => c + 1,
        };
        self.val = self.val & 0xFF0F | ((val as u16) << 4);
    }
}

#[derive(Debug)]
pub struct Cell {
    pub ch: char,
    pub attr: Attr,
}

pub struct VT {
    pub lines: Vec<Box<Vec<Cell>>>,
    /// The number of lines visible on screen.
    pub height: usize,
    /// The index of the first visible line on screen.
    pub top: usize,
    /// The index of the line the cursor is on, relative to the lines buffer.
    pub row: usize,
    /// The column the cursor is on.
    pub col: usize,
    /// The current display attributes, to be used when text is added.
    pub attr: Attr,
}

impl VT {
    pub fn new() -> VT {
        VT {
            lines: Vec::new(),
            height: 24,
            top: 0,
            row: 0,
            col: 0,
            attr: Attr { val: 0 },
        }
    }

    pub fn ensure_pos(&mut self) -> &mut Cell {
        while self.row >= self.lines.len() {
            self.lines.push(Box::new(Vec::new()));
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
    r: ByteScanner<'a, fs::File>,
    w: fs::File,
}

impl<'a> VTReader<'a> {
    pub fn new(vt: &'a Mutex<VT>, r: &'a mut fs::File, w: fs::File) -> VTReader<'a> {
        VTReader {
            todo: HashSet::new(),
            vt: vt,
            r: ByteScanner::new(r),
            w: w,
        }
    }

    fn todo<S: Into<String> + Display>(&mut self, msg: S) {
        let msg = msg.into();
        if self.todo.insert(msg.clone()) {
            println!("TODO: {}", msg)
        }
    }

    fn read_num(&mut self) -> usize {
        let mut num = 0;
        loop {
            let c = self.r.next().unwrap();
            match c as char {
                '0'...'9' => num = num * 10 + ((c - ('0' as u8)) as usize),
                _ => {
                    self.r.back();
                    return num;
                }
            }
        }
    }

    fn read_csi(&mut self) {
        let mut question = false;
        let mut gt = false;

        let mut args: [usize; 2] = [0; 2];
        let mut argc = 0;
        loop {
            match self.r.next().unwrap() as char {
                '?' => question = true,
                '>' => gt = true,
                '0'...'9' => {
                    self.r.back();
                    if argc == args.len() {
                        panic!("too many args")
                    }
                    args[argc] = self.read_num();
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

        let cmd = self.r.next().unwrap() as char;
        match cmd {
            'c' if gt => {
                // send device attributes (secondary)
                self.w.write("\x1b[41;0;0c".as_bytes()).unwrap();
            }
            'h' | 'l' if question => {
                // set = cmd == 'h'
                match args[0] {
                    1 => self.todo("application cursor keys mode"),
                    12 => self.todo("start blinking cursor"),
                    25 => self.todo("show/hide cursor"),
                    1049 => {
                        self.todo("save cursor");
                        self.todo("alternate screen buffer");
                    }
                    arg => self.todo(format!("?h arg {}", arg)),
                }
            }
            'l' if question => {
                self.todo("dec private mode");
            }
            'm' => {
                // character attributes
                let mut vt = self.vt.lock().unwrap();
                if args.len() == 0 {
                    vt.attr.val = 0;
                }
                for attr in args {
                    match *attr {
                        0 => vt.attr.val = 0,
                        1 => vt.attr.set_bold(),
                        7 => vt.attr.set_inverse(true),
                        27 => vt.attr.set_inverse(false),
                        v if v >= 30 && v < 39 => vt.attr.set_fg(Some(v - 30)),
                        39 => vt.attr.set_fg(None),
                        v if v >= 40 && v < 49 => vt.attr.set_bg(Some(v - 30)),
                        49 => vt.attr.set_bg(None),
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
                let vt = self.vt.lock().unwrap();
                let top = *args.get(0).unwrap_or(&1);
                let bottom = *args.get(1).unwrap_or(&vt.height);
                if top == 1 && bottom == vt.height {
                    // full window
                } else {
                    self.todo(format!("set scrolling region: {}:{}", top, bottom));
                }
            }
            'A' => {
                // cursor up
                let dy = *args.get(0).unwrap_or(&1);
                let mut vt = self.vt.lock().unwrap();
                if vt.row > dy {
                    vt.row -= dy;
                }
            }
            'C' => {
                // cursor forward
                let dx = *args.get(0).unwrap_or(&1);
                let mut vt = self.vt.lock().unwrap();
                vt.col += dx;
            }
            'D' => {
                // cursor back
                let dx = *args.get(0).unwrap_or(&1);
                let mut vt = self.vt.lock().unwrap();
                if vt.col > dx {
                    vt.col -= dx;
                }
            }
            'H' => {
                let mut vt = self.vt.lock().unwrap();
                let row = *args.get(0).unwrap_or(&1) - 1;
                vt.row = vt.top + row;
                vt.col = *args.get(1).unwrap_or(&1) - 1;
            }
            'J' => {
                let mut vt = self.vt.lock().unwrap();
                match *args.get(0).unwrap_or(&0) {
                    2 => {
                        let top = vt.top;
                        vt.lines.truncate(top + 1);
                        vt.row = top;
                        vt.col = 0;
                    }
                    x => self.todo(format!("erase in display {}", x)),
                }
            }
            'K' => {
                match *args.get(0).unwrap_or(&0) {
                    0 => {
                        // Erase to Right.
                        let mut vt = self.vt.lock().unwrap();
                        vt.ensure_pos();
                        let row = vt.row;
                        let col = vt.col;
                        vt.lines[row].truncate(col);
                    }
                    mode => {
                        self.todo(format!("erase in line {:?}", mode));
                    }
                }
            }
            c => panic!("unhandled CSI {}", c),
        }
    }

    fn read_osc(&mut self) {
        let cmd = self.read_num();
        match self.r.next().unwrap() as char {
            ';' => {}
            _ => panic!("unexpected OSC"),
        }
        let mut text = String::new();
        for c in &mut self.r {
            match c {
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
    }

    fn read_escape(&mut self) {
        match self.r.next().unwrap() as char {
            '(' => {
                match self.r.next().unwrap() as char {
                    'B' => {
                        // US ASCII
                    }
                    c => self.todo(format!("g0 charset {}", c)),
                }
            }
            '=' => self.todo("application keypad"),
            '[' => self.read_csi(),
            ']' => self.read_osc(),
            '>' => self.todo("normal keypad"),
            c => panic!("notimpl: esc {}", c),
        }
    }

    fn read_utf8(&mut self) -> u32 {
        let c = self.r.next().unwrap();
        let (n, mut rune) = {
            if c & 0b11100000 == 0b11000000 {
                (1, (c & 0b00011111) as u32)
            } else if c & 0b11110000 == 0b11100000 {
                (2, (c & 0b00001111) as u32)
            } else {
                return '?' as u32;
            }
        };

        for _ in 0..n {
            let c = self.r.next().unwrap();
            if c & 0b11000000 == 0b10000000 {
                rune = rune << 6 | (c & 0b00111111) as u32;
            } else {
                self.r.back();
                return '?' as u32;
            }
        }
        rune
    }

    pub fn read(&mut self) -> bool {
        let c = match self.r.next() {
            None => return false,
            Some(c) => c,
        };

        match c as char {
            '\x07' => self.todo("bell"),
            '\x08' => {
                let mut vt = self.vt.lock().unwrap();
                if vt.col > 0 {
                    vt.col -= 1;
                }
            }
            '\x09' => {
                let mut vt = self.vt.lock().unwrap();
                vt.col += 8 - (vt.col % 8);
            }
            '\n' => {
                let mut vt = self.vt.lock().unwrap();
                vt.row += 1;
                vt.col = 0;
            }
            '\r' => {
                let mut vt = self.vt.lock().unwrap();
                vt.col = 0;
            }
            '\x1b' => self.read_escape(),
            c if c >= ' ' && c <= '~' => {
                let mut vt = self.vt.lock().unwrap();
                *vt.ensure_pos() = Cell {
                    ch: c,
                    attr: vt.attr.clone(),
                };
                vt.col += 1
            }
            c if c as u8 >= 0x80 => {
                self.r.back();
                let rune = self.read_utf8();
                println!("rune {}", rune);
            }
            _ => {
                panic!("unhandled input {:?}", c);
            }
        }
        return true;
    }
}
