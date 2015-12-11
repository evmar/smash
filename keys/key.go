// Package keys abstracts keyboard input, for use both in mapping X
// events to key events and in readline's key handling.
package keys

import "fmt"

//go:generate stringer -type=Sym

// Sym is a key symbol: e.g. the 'a' key.
type Sym int

const (
	NoSym Sym = 0

	// 1 through 127 are the same as in ASCII.

	Backspace Sym = 8
	Tab       Sym = 9
	Enter     Sym = 13
	Esc       Sym = 27

	Left Sym = iota + 128
	Right
	Up
	Down

	FirstNonASCIISym = Left
)

const (
	ModControl uint = 1 << iota
	ModMeta
)

type Key struct {
	Sym Sym

	// Modifier mask: Control, Meta, etc.
	Mods uint
}

func (sym Sym) IsText() bool {
	return sym >= ' ' && sym <= '~'
}

func (sym Sym) Name() string {
	if sym.IsText() {
		return fmt.Sprintf("%c", sym)
	}
	return sym.String()
}

// Spec returns a string specifying the key, e.g. "C-M-a" for Control+Meta+a.
func (k Key) Spec() string {
	spec := ""
	if k.Mods&ModControl != 0 {
		spec += "C-"
	}
	if k.Mods&ModMeta != 0 {
		spec += "M-"
	}
	spec += k.Sym.Name()
	return spec
}
