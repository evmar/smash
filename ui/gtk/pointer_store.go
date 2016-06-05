// +build !headless

package gtk

import "unsafe"

// CKey is the type of opaque tokens used to identify objects that
// round-trip through C.
type CKey uintptr

// ViaC is the interface implemented by objects that have a CKey.
type ViaC interface {
	// CKey returns a place to store the object's specific CKey.
	CKey() *CKey
}

// pointerStore manages mapping from CKey<->object.
type pointerStore struct {
	// We're only used from one thread, so no synchronization needed.
	// sync.Mutex
	ptrs map[CKey]ViaC
	next CKey
}

// Key gets the CKey associated with a ViaC, initiailizing it if necessary.
func (ps *pointerStore) Key(obj ViaC) unsafe.Pointer {
	key := obj.CKey()
	if *key == 0 {
		// ps.Lock()
		ps.next++
		*key = ps.next
		ps.ptrs[*key] = obj
		// ps.Unlock()
	}
	return unsafe.Pointer(*key)
}

// Get gets the object associated with a CKey.
func (ps *pointerStore) Get(key unsafe.Pointer) ViaC {
	// ps.Lock(); defer ps.Unlock()
	return ps.ptrs[CKey(key)]
}

// Remove drops an object from the pointerStore.
func (ps *pointerStore) Remove(key unsafe.Pointer) {
	// ps.Lock(); defer ps.Unlock()
	obj := ps.Get(key)
	*obj.CKey() = 0
	delete(ps.ptrs, CKey(key))
}

var globalPointerStore = pointerStore{
	ptrs: make(map[CKey]ViaC),
}
