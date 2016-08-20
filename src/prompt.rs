extern crate cairo;
extern crate gdk;
use std::rc::Rc;
use view;
use view::Layout;
use readline::ReadLineView;

pub struct Prompt {
    rl: Rc<ReadLineView>,
}

impl Prompt {
    pub fn new(dirty: Rc<Fn()>) -> Prompt {
        Prompt { rl: ReadLineView::new(dirty) }
    }

    pub fn set_accept_cb(&self, accept_cb: Box<FnMut(&str)>) {
        self.rl.rl.borrow_mut().accept_cb = accept_cb;
    }
}

impl view::View for Prompt {
    fn draw(&self, cr: &cairo::Context, focus: bool) {
        cr.save();
        cr.set_source_rgb(0.7, 0.7, 0.7);
        cr.new_path();
        cr.move_to(5.0, 8.0);
        let height = self.get_layout().height as f64;
        cr.line_to(13.0, height / 2.0);
        cr.line_to(5.0, height - 8.0);
        cr.fill();

        cr.translate(18.0, 5.0);
        self.rl.draw(cr, focus);
        cr.restore();
    }
    fn key(&self, ev: &gdk::EventKey) {
        self.rl.key(ev);
    }

    fn relayout(&self, cr: &cairo::Context, space: Layout) -> Layout {
        self.rl.relayout(cr, space.add(-20, -10));
        self.get_layout()
    }
    fn get_layout(&self) -> Layout {
        self.rl.get_layout().add(20, 10)
    }
}
