package main

import (
	"log/slog"
	"time"

	"github.com/abenz1267/elephant/internal/common"
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
	config     *Config
)

type Config struct {
	common.Config           `koanf:",squash"`
	LaunchPrefix            string `koanf:"launch_prefix" desc:"overrides the default app2unit or uwsm prefix, if set. 'CLEAR' to not prefix." default:""`
	Locale                  string `koanf:"locale" desc:"to override systems locale" default:""`
	ShowActions             bool   `koanf:"show_actions" desc:"include application actions, f.e. 'New Private Window' for Firefox" default:"false"`
	ShowGeneric             bool   `koanf:"show_generic" desc:"include generic info when show_actions is true" default:"false"`
	ShowActionsWithoutQuery bool   `koanf:"show_actions_without_query" desc:"show application actions, if the search query is empty" default:"false"`
	History                 bool   `koanf:"history" desc:"make use of history for sorting" default:"false"`
}

func init() {
	start := time.Now()
	config = &Config{
		Config:                  common.Config{},
		ShowActions:             false,
		ShowGeneric:             false,
		ShowActionsWithoutQuery: false,
		History:                 false,
	}

	common.LoadConfig(Name, config)

	loadFiles()

	slog.Info(Name, "desktop files", len(files), "time", time.Since(start))
}
