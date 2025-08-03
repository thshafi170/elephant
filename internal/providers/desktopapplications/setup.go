package main

import (
	"github.com/abenz1267/elephant/internal/common/history"
)

type DesktopFile struct {
	Data
	Actions []Data
}

var (
	Name       = "desktopapplications"
	NamePretty = "Desktop Applications"
	h          = history.Load(Name)
)

func init() {
	loadConfig()
}

func Load() {
	loadFiles()
}
