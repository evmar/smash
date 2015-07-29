package keys

import "fmt"

//go:generate stringer -type=Sym
type Sym int

const (
	SymNone Sym = 0

	// 1 through 127 as in ASCII.
	Backspace Sym = 8
	Tab       Sym = 9
	NL        Sym = 10
	CR        Sym = 13

	Left Sym = iota + 128
	Right
	Up
	Down

	SymFirstNonASCII = Left
)

const (
	ModControl uint = 1 << iota
	ModMeta
)

type Key struct {
	Sym  Sym
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
