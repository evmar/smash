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
    pub fn parse(&self, cmd: &str) -> Command {
        let argv = cmd.split(' ').map(String::from).collect();
        Command::External(argv)
    }
}
