// Package providers provides common provider functions.
package providers

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"plugin"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/charlievieth/fastwalk"
)

type Provider struct {
	Name        *string
	NamePretty  *string
	Load        func()
	PrintConfig func()
	Cleanup     func()
	Activate    func(sid uint32, identifier, action string)
	Query       func(text string) []common.Entry
}

var Providers map[string]Provider

func Load() {
	Providers = make(map[string]Provider)
	dir := filepath.Join(common.ConfigDir(), "providers")

	if !common.FileExists(dir) {
		slog.Error("elephant", "providers", "you don't have any providers installed")
		os.Exit(1)
	}

	conf := fastwalk.Config{
		Follow: true,
	}

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("providers", "load", err)
			os.Exit(1)
		}

		if filepath.Ext(path) == ".so" {
			p, err := plugin.Open(path)
			if err != nil {
				panic(err)
			}

			name, err := p.Lookup("Name")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			namePretty, err := p.Lookup("NamePretty")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			loadFunc, err := p.Lookup("Load")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			activateFunc, err := p.Lookup("Activate")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			queryFunc, err := p.Lookup("Query")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			cleanupFunc, err := p.Lookup("Cleanup")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			printDocFunc, err := p.Lookup("PrintDoc")
			if err != nil {
				slog.Error("providers", "load", err, "provider", path)
			}

			provider := Provider{
				Load:        loadFunc.(func()),
				Name:        name.(*string),
				Cleanup:     cleanupFunc.(func()),
				Activate:    activateFunc.(func(sid uint32, identifier, action string)),
				Query:       queryFunc.(func(string) []common.Entry),
				NamePretty:  namePretty.(*string),
				PrintConfig: printDocFunc.(func()),
			}

			Providers[*provider.Name] = provider
		}

		return err
	}

	if err := fastwalk.Walk(&conf, dir, walkFn); err != nil {
		slog.Error("providers", "load", err)
		os.Exit(1)
	}

	if len(Providers) == 0 {
		slog.Error("elephant", "providers", "you don't have any providers installed")
		os.Exit(1)
	}
}

func Setup() {
	for _, v := range Providers {
		v.Load()
	}
}
