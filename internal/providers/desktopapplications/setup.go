package main

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/charlievieth/fastwalk"
)

type DesktopFile struct {
	Data
	Actions []Data
}

var files map[string]*DesktopFile

var (
	Name       = "desktopapplications"
	NamePretty = "Desktop Applications"
)

// TODO: watch folders for changes
func Load() {
	loadConfig()

	start := time.Now()
	files = make(map[string]*DesktopFile)

	ll := config.Locale

	if ll == "" {
		ll = os.Getenv("LANG")

		langMessages := os.Getenv("LC_MESSAGES")
		if langMessages != "" {
			ll = langMessages
		}

		langAll := os.Getenv("LC_ALL")
		if langAll != "" {
			ll = langAll
		}

		ll = strings.Split(ll, ".")[0]
	}

	l := strings.Split(ll, "_")[0]

	dirs := xdg.ApplicationDirs

	conf := fastwalk.Config{
		Follow: true,
	}

	for _, root := range dirs {
		if _, err := os.Stat(root); err != nil {
			continue
		}

		walkFn := func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				slog.Error(Name, "walk", err)
				os.Exit(1)
			}

			if _, ok := files[path]; ok {
				return nil
			}

			if !d.IsDir() && filepath.Ext(path) == ".desktop" {
				files[path] = parseFile(path, l, ll)
			}

			return err
		}

		if err := fastwalk.Walk(&conf, root, walkFn); err != nil {
			slog.Error(Name, "walk", err)
			os.Exit(1)
		}
	}

	slog.Info(Name, "files", len(files), "time", time.Since(start))
}
