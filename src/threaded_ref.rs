extern crate thread_id;

/// ThreadedRef is a reference to an object that lives only on one thread.
/// When it's created it takes ownership of T and saves which thread it's
/// on; later attempts to access the T will panic unless they are done on
/// the original thread.
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
