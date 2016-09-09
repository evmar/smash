use std::path::PathBuf;

pub struct Shell {
    cwd: PathBuf,
}

pub enum Command {
    Builtin(fn(&mut Shell)),
    External(Vec<String>),
}

impl Shell {
    pub fn new() -> Shell {
        Shell { cwd: PathBuf::from("/") }
    }
}

pub fn parse(cmd: &str) -> Command {
    let argv: Vec<_> = cmd.split(' ').map(String::from).collect();
    // Note: split() always returns at least one element.
    match argv[0].as_str() {
        "" => {
            fn none(_: &mut Shell) {}
            Command::Builtin(none)
        }
        _ => Command::External(argv),
    }
}
