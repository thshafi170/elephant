// Package providers provides common provider functions.
package providers

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/charlievieth/fastwalk"
)

type Provider struct {
	Name          *string
	NamePretty    *string
	Load          func()
	PrintDoc      func()
	Cleanup       func(qid uint32)
	EntryToString func(common.Entry) string
	Activate      func(qid uint32, identifier, action string)
	Query         func(qid uint32, text string) []common.Entry
}

var Providers map[string]Provider

func Load() {
	start := time.Now()
	Providers = make(map[string]Provider)
	dir := filepath.Join(common.ConfigDir(), "providers")

	if !common.FileExists(dir) {
		slog.Error("providers", "load", "you don't have any providers installed")
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

			entryToStringFunc, err := p.Lookup("EntryToString")
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
				Load:          loadFunc.(func()),
				Name:          name.(*string),
				EntryToString: entryToStringFunc.(func(common.Entry) string),
				Cleanup:       cleanupFunc.(func(uint32)),
				Activate:      activateFunc.(func(qid uint32, identifier, action string)),
				Query:         queryFunc.(func(uint32, string) []common.Entry),
				NamePretty:    namePretty.(*string),
				PrintDoc:      printDocFunc.(func()),
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
		slog.Error("providers", "load", "you don't have any providers installed")
		os.Exit(1)
	}

	slog.Info("providers", "loaded", len(Providers), "time", time.Since(start))
}

func Setup() {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(Providers))

	for _, v := range Providers {
		go func(wg *sync.WaitGroup, p Provider) {
			defer wg.Done()
			p.Load()
		}(&wg, v)
	}

	wg.Wait()

	slog.Info("providers", "setup", time.Since(start))
}
