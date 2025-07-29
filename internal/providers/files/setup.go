package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/internal/util"
	"github.com/adrg/xdg"
	"github.com/charlievieth/fastwalk"
)

var (
	paths        []string
	results      providers.QueryData[[]string]
	resultsMutex sync.Mutex
	terminal     string
)

var terminalApps map[string]struct{}

var (
	Name       = "files"
	NamePretty = "Files"
)

func init() {
	loadConfig()
	results = providers.QueryData[[]string]{}
	terminalApps = map[string]struct{}{}
}

func PrintDoc() {
	fmt.Printf("### %s\n", Name)
	fmt.Println("Search files and folders.")
	fmt.Println()
	util.PrintConfig(Config{})
}

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)
	resultsMutex.Lock()
	delete(results.Queries, qid)
	resultsMutex.Unlock()
}

func findTerminalApps() {
	conf := fastwalk.Config{
		Follow: true,
	}

	for _, root := range xdg.ApplicationDirs {
		if _, err := os.Stat(root); err != nil {
			continue
		}

		if err := fastwalk.Walk(&conf, root, func(path string, d fs.DirEntry, err error) error {
			if strings.HasSuffix(path, ".desktop") {
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				if bytes.Contains(b, []byte("Terminal=true")) {
					terminalApps[filepath.Base(path)] = struct{}{}
				}
			}
			return nil
		}); err != nil {
			slog.Error(Name, "walk", err)
			os.Exit(1)
		}
	}
}

func Load() {
	start := time.Now()

	findTerminalApps()
	terminal = common.GetTerminal()

	paths = []string{}
	home, _ := os.UserHomeDir()
	cmd := exec.Command("fd", ".", home, "--ignore-vcs", "--type", "file", "--type", "directory")

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error(Name, "files", err)
		os.Exit(1)
	}

	for v := range bytes.Lines(out) {
		if len(v) > 0 {
			paths = append(paths, strings.TrimSpace(string(v)))
		}
	}

	slog.Info(Name, "files", len(paths), "time", time.Since(start))
}

func EntryToString(e common.Entry) string {
	var start int
	var field string

	positions := []string{}

	if e.Fuzzy != nil {
		if e.Fuzzy.Pos != nil {
			for _, num := range *e.Fuzzy.Pos {
				positions = append(positions, strconv.Itoa(num))
			}
		}

		start = e.Fuzzy.Start
		field = e.Fuzzy.Field
	}

	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;%d;%s", e.Provider, e.Identifier, e.Text, "", "", strings.Join(positions, ","), start, field)
}
