extern crate cairo;
extern crate gdk;
use std::rc::Rc;
use std::cell::RefCell;
use view;
use view::Layout;
use readline::ReadLineView;

pub struct Prompt {
    rl: Rc<RefCell<ReadLineView>>,
}

impl Prompt {
    pub fn new(dirty: Rc<Fn()>) -> Prompt {
        Prompt { rl: ReadLineView::new(dirty) }
    }

    pub fn set_accept_cb(&mut self, accept_cb: Box<FnMut(&str)>) {
        self.rl.borrow_mut().rl.accept_cb = accept_cb;
    }
}

impl view::View for Prompt {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        let gray = 0.3;
        let x_pad = 5.0;
        let y_pad = 10.0;
        let width = 8.0;
        cr.save();
        cr.translate(x_pad, 0.0);
        cr.set_source_rgb(gray, gray, gray);
        cr.new_path();
        cr.move_to(0.0, y_pad);
        let height = self.get_layout().height as f64;
        cr.line_to(width, height / 2.0);
        cr.line_to(0.0, height - y_pad);
        cr.fill();

        cr.translate(width + x_pad, 5.0);
        self.rl.borrow_mut().draw(cr, focus);
        cr.restore();
    }
    fn key(&mut self, ev: &gdk::EventKey) {
        self.rl.borrow_mut().key(ev);
    }

    fn relayout(&mut self, cr: &cairo::Context, space: Layout) -> Layout {
        self.rl.borrow_mut().relayout(cr, space.add(-20, -10));
        self.get_layout()
    }
    fn get_layout(&self) -> Layout {
        self.rl.borrow_mut().get_layout().add(20, 10)
    }
}
