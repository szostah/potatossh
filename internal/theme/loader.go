package theme

import (
	"encoding/json"
	"fmt"
	"os"
)

var THEMES = []string{"dracula.json", "gruvbox.json", "nord.json", "solarized-dark.json", "solarized-light.json"}

func Load() []Theme {
	list := []Theme{}
	for _, t := range THEMES {
		fileBytes, err := os.ReadFile("assets/themes/" + t)
		if err != nil {
			fmt.Printf("Can not load theme %s: %s.\n", t, err)
			continue
		}
		var theme Theme
		json.Unmarshal(fileBytes, &theme)
		list = append(list, theme)
	}
	return list
}
