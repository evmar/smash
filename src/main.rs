extern crate smash;
use smash::view;

fn main() {
    view::init();
    let win = view::Win::new();

    {
        let mut win = win.borrow_mut();
        win.show();
    }

    view::main();
}
