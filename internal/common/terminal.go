package common

import (
	"os"
	"os/exec"
)

func GetTerminal() string {
	envVars := []string{"TERM", "TERMINAL"}

	for _, v := range envVars {
		term, ok := os.LookupEnv(v)
		if ok {
			path, _ := exec.LookPath(term)

			if path != "" {
				return path
			}
		}
	}

	t := []string{
		"kitty",
		"foot",
		"ghostty",
		"alacritty",
		"Eterm",
		"aterm",
		"gnome-terminal",
		"guake",
		"hyper",
		"konsole",
		"lilyterm",
		"lxterminal",
		"mate-terminal",
		"qterminal",
		"roxterm",
		"rxvt",
		"st",
		"terminator",
		"terminix",
		"terminology",
		"termit",
		"termite",
		"tilda",
		"tilix",
		"urxvt",
		"uxterm",
		"wezterm",
		"x-terminal-emulator",
		"xfce4-terminal",
		"xterm",
	}

	for _, v := range t {
		path, _ := exec.LookPath(v)

		if path != "" {
			return path
		}
	}

	return ""
}
