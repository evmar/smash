extern crate thread_id;

use std::mem;

/// ThreadedRef is a reference to an object that lives only on one thread.
/// When it's created it takes ownership of T and saves which thread it's
/// on; later attempts to access the T will panic unless they are done on
/// the original thread.
pub struct ThreadedRef<T> {
    tid: usize,
    data: Option<T>,
}

unsafe impl<T> Send for ThreadedRef<T> {}
unsafe impl<T> Sync for ThreadedRef<T> {}

impl<T> ThreadedRef<T> {
    pub fn new(data: T) -> ThreadedRef<T> {
        ThreadedRef {
            tid: thread_id::get(),
            data: Some(data),
        }
    }

    pub fn check(&self) {
        assert_eq!(self.tid, thread_id::get());
    }

    pub fn get(&self) -> &T {
        self.check();
        match self.data {
            Some(ref data) => data,
            None => panic!("no data"),
        }
    }

    pub fn take(&mut self) -> T {
        self.check();
        mem::replace(&mut self.data, None).unwrap()
    }
}
