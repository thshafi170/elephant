package main

type DesktopFile struct {
	Data
	Actions []Data
}

var (
	Name       = "desktopapplications"
	NamePretty = "Desktop Applications"
)

func init() {
	loadConfig()
}

func Load() {
	loadFiles()
}
