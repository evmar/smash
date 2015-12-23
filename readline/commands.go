package readline

import (
	"fmt"
	"unicode"

	"smash/keys"
)

type Command func(rl *ReadLine, key keys.Key)

func isWordChar(c byte) bool {
	// All this code should be rune-based anyway...
	return unicode.IsLetter(rune(c))
}

func backwardWord(rl *ReadLine, key keys.Key) {
	for rl.Pos > 0 && !isWordChar(rl.Text[rl.Pos-1]) {
		rl.Pos--
	}
	for rl.Pos > 0 && isWordChar(rl.Text[rl.Pos-1]) {
		rl.Pos--
	}
}

var commands = map[string]Command{
	// Moving
	"beginning-of-line": func(rl *ReadLine, key keys.Key) {
		rl.Pos = 0
	},
	"end-of-line": func(rl *ReadLine, key keys.Key) {
		rl.Pos = len(rl.Text)
	},
	"forward-char": func(rl *ReadLine, key keys.Key) {
		if rl.Pos < len(rl.Text) {
			rl.Pos++
		}
	},
	"backward-char": func(rl *ReadLine, key keys.Key) {
		if rl.Pos > 0 {
			rl.Pos--
		}
	},
	"forward-word": func(rl *ReadLine, key keys.Key) {
		// TODO: make this behavior make sense?
		for rl.Pos < len(rl.Text) && !isWordChar(rl.Text[rl.Pos]) {
			rl.Pos++
		}
		for rl.Pos < len(rl.Text) && isWordChar(rl.Text[rl.Pos]) {
			rl.Pos++
		}
	},
	"backward-word": backwardWord,

	// History
	"accept-line": func(rl *ReadLine, key keys.Key) {
		rl.Pos = -1
		if rl.Accept(rl.String()) {
			rl.Config.History = append(rl.Config.History, rl.String())
			rl.Config.HistoryPos = len(rl.Config.History)
		}
	},
	"previous-history": func(rl *ReadLine, key keys.Key) {
		if rl.Config.HistoryPos == 0 {
			return
		}
		rl.Config.HistoryPos--
		rl.Text = []byte(rl.Config.History[rl.Config.HistoryPos])
		rl.Pos = len(rl.Text)
	},
	"next-history": func(rl *ReadLine, key keys.Key) {
		if rl.Config.HistoryPos+1 == len(rl.Config.History) {
			return
		}
		rl.Config.HistoryPos++
		rl.Text = []byte(rl.Config.History[rl.Config.HistoryPos])
		rl.Pos = len(rl.Text)
	},

	// Text
	"backward-delete-char": func(rl *ReadLine, key keys.Key) {
		if rl.Pos == 0 {
			return
		}
		copy(rl.Text[rl.Pos-1:], rl.Text[rl.Pos:])
		rl.Text = rl.Text[:len(rl.Text)-1]
		rl.Pos--
	},
	"self-insert": func(rl *ReadLine, key keys.Key) {
		rl.Insert(byte(key.Sym))
	},

	// Killing
	"kill-line": func(rl *ReadLine, key keys.Key) {
		rl.Text = rl.Text[:rl.Pos]
	},
	"kill-word": func(rl *ReadLine, key keys.Key) {
		pos := rl.Pos
		backwardWord(rl, key)
		rl.Text = append(rl.Text[:rl.Pos], rl.Text[pos:]...)
	},
	"unix-line-discard": func(rl *ReadLine, key keys.Key) {
		copy(rl.Text, rl.Text[rl.Pos:])
		rl.Text = rl.Text[:len(rl.Text)-rl.Pos]
		rl.Pos = 0
	},

	// Completion
	"complete": func(rl *ReadLine, key keys.Key) {
		rl.StartComplete()
	},
}

func DefaultBindings() map[string]string {
	b := map[string]string{
		// Moving
		"C-a": "beginning-of-line",
		"C-e": "end-of-line",
		"C-f": "forward-char",
		"C-b": "backward-char",
		"M-f": "forward-word",
		"M-b": "backward-word",

		"Right": "forward-char",
		"Left":  "backward-char",

		// History
		"Enter": "accept-line",
		"C-p":   "previous-history",
		"C-n":   "next-history",

		// Text
		"C-h":       "backward-delete-char",
		"Backspace": "backward-delete-char",

		// Killing
		"C-k":         "kill-line",
		"M-Backspace": "kill-word",
		"C-u":         "unix-line-discard",

		// Completion
		"Tab": "complete",
	}
	for ch := ' '; ch <= '~'; ch++ {
		b[fmt.Sprintf("%c", ch)] = "self-insert"
	}
	return b
}
