package smash

import (
	"github.com/evmar/smash/readline"
	"github.com/evmar/smash/shell"
	"github.com/evmar/smash/ui/fake"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPromptDelegate struct {
}

func (tpd *testPromptDelegate) OnPromptAccept(string) bool {
	panic("x")
}

func (tpd *testPromptDelegate) GetPromptAbsolutePosition(pv *PromptView) (int, int) {
	return 0, 0
}

func (tpd *testPromptDelegate) Complete(input string) (int, []string, error) {
	switch input {
	case "":
		// No input => no completions.
		return 0, []string{}, nil
	case "ls l":
		return 3, []string{"log", "logview.go"}, nil
	default:
		panic("x")
	}
}

func (tpd *testPromptDelegate) Chdir(input string) error {
	return nil
}

func TestFilter(t *testing.T) {
	text, comps := filterPrefix("", []string{})
	assert.Equal(t, text, "")
	assert.Equal(t, comps, []string{})

	text, comps = filterPrefix("foo", []string{})
	assert.Equal(t, text, "foo")
	assert.Equal(t, comps, []string{})

	text, comps = filterPrefix("l", []string{"foo", "log", "logview.go"})
	assert.Equal(t, text, "log")
	assert.Equal(t, comps, []string{"log", "logview.go"})
}

func TestComplete(t *testing.T) {
	ui := fake.NewUI()
	parent := NewTestViewHost(ui)
	delegate := &testPromptDelegate{}
	config := &readline.Config{}
	shell := shell.NewShell(".", map[string]string{}, delegate)
	pv := NewPromptView(parent, delegate, config, shell)

	pv.StartComplete()
	ui.RunQueue(true)
	assert.Equal(t, pv.readline.Text, []byte(""))

	pv.readline.Text = []byte("ls l")
	pv.readline.Pos = 4
	pv.StartComplete()
	ui.RunQueue(true)
	assert.Equal(t, pv.readline.Text, []byte("ls log"))
}
