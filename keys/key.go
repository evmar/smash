package keys

import "fmt"

//go:generate stringer -type=KeySym
type KeySym int

const (
	KeyNone KeySym = 0

	// 1 through 127 as in ASCII.
	KeyBackspace KeySym = 8
	KeyTab       KeySym = 9
	KeyNL        KeySym = 10
	KeyCR        KeySym = 13

	KeyLeft KeySym = iota + 128
	KeyRight
	KeyUp
	KeyDown

	KeyFirstNonASCII = KeyLeft
)

const (
	KeyModControl uint = 1 << iota
	KeyModMeta
)

type Key struct {
	Sym  KeySym
	Mods uint
}

func (sym KeySym) IsText() bool {
	return sym >= ' ' && sym <= '~'
}

func (sym KeySym) Name() string {
	if sym.IsText() {
		return fmt.Sprintf("%c", sym)
	}
	return sym.String()
}

func (k Key) Spec() string {
	spec := ""
	if k.Mods&KeyModControl != 0 {
		spec += "C-"
	}
	if k.Mods&KeyModMeta != 0 {
		spec += "M-"
	}
	spec += k.Sym.Name()
	return spec
}
