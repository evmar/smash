package main

/*
#cgo pkg-config: gtk+-3.0
#cgo LDFLAGS: libsmashgtk.a
#include "smashgtk.h"
*/
import "C"

func main() {
	C.run()
}
