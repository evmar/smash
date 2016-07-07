extern crate thread_id;

pub struct ThreadedRef<T> {
    tid: usize,
    data: T,
}

unsafe impl<T> Send for ThreadedRef<T> {}
unsafe impl<T> Sync for ThreadedRef<T> {}

impl<T> ThreadedRef<T> {
    pub fn new(data: T) -> ThreadedRef<T> {
        ThreadedRef {
            tid: thread_id::get(),
            data: data,
        }
    }

    pub fn check(&self) {
        assert_eq!(self.tid, thread_id::get());
    }

    pub fn get(&self) -> &T {
        self.check();
        &self.data
    }
}
