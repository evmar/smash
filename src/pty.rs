extern crate libc;

use std::ffi::CString;
use std::fs;
use std::os::unix::io::FromRawFd;
use std::os::unix::io::AsRawFd;
use std::ptr;

pub fn bash() -> (fs::File, fs::File) {
    let args = ["bash\0"];
    let mut cargs: Vec<_> = args.iter().map(|&a| a.as_ptr() as *const i8).collect();
    cargs.push(ptr::null());

    let mut fd: libc::c_int = 0;
    unsafe {
        let pid = libc::forkpty(&mut fd, ptr::null_mut(), ptr::null(), ptr::null());
        match pid {
            -1 => panic!("forkpty failed"),
            0 => {
                // child
                if libc::execvp(cargs[0], cargs.as_ptr()) < 0 {
                    libc::perror(CString::new("exec").unwrap().as_ptr());
                }
                panic!("notreached");
            }
            _ => {
                return (fs::File::from_raw_fd(fd), fs::File::from_raw_fd(fd));
            }
        }
    }
}

pub fn set_size(f: &fs::File, rows: u16, cols: u16) {
    let winsize = libc::winsize {
        ws_row: rows,
        ws_col: cols,
        ws_xpixel: 0,
        ws_ypixel: 0,
    };
    let err = unsafe { libc::ioctl(f.as_raw_fd(), libc::TIOCSWINSZ, &winsize) };
    if err < 0 {
        println!("TIOCSWINSZ {:?}", err);
    }
}
