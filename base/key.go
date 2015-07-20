package base

type KeySym int

const (
	KeyNone KeySym = iota
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
)

type Key struct {
	Text    string
	Special KeySym
}
