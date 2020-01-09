package bash

import "testing"

func TestParseAliases(t *testing.T) {
	aliases, err := parseAliases("alias foo='bar'\nalias bar='baz'\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(aliases) != 2 {
		t.Errorf("want 2 aliases, got %q", aliases)
	}
	if aliases["foo"] != "bar" || aliases["bar"] != "baz" {
		t.Errorf("wanted foo=bar, bar=baz aliases, got %q", aliases)
	}
}
