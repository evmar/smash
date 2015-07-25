package readline

type Command func(rl *ReadLine, key Key)

var commands = map[string]Command{
	"beginning-of-line": func(rl *ReadLine, key Key) {
		rl.Pos = 0
	},
	"self-insert": func(rl *ReadLine, key Key) {
		rl.Insert(byte(key.Ch))
	},
}
