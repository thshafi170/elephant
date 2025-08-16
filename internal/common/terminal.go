package common

import (
	"fmt"
	"os"
	"os/exec"
)

var terminal = ""

func init() {
	terminal = GetTerminal()
}

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

func WrapWithTerminal(in string) string {
	if terminal == "" {
		return in
	}

	return fmt.Sprintf("%s %s", terminal, in)
}
