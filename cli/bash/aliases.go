package bash

import (
	"fmt"
	"os/exec"
	"regexp"
)

func parseAliases(text string) (map[string]string, error) {
	if text == "" {
		return map[string]string{}, nil
	}
	re := regexp.MustCompile(`(?m)^alias (\w+)='(.*?)'\n`)
	match := re.FindAllStringSubmatch(text, -1)
	if match == nil {
		return nil, fmt.Errorf("couldn't parse aliases %q", text)
	}
	aliases := map[string]string{}
	for _, row := range match {
		if len(row) != 3 {
			return nil, fmt.Errorf("bad alias output %q", row)
		}
		aliases[row[1]] = row[2]
	}
	return aliases, nil
}

// GetAliases shells out to bash to extract the user's configured alias list.
func GetAliases() (map[string]string, error) {
	cmd := exec.Command("bash", "-i", "-c", "alias")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseAliases(string(out))
}
