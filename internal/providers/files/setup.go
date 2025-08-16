package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/internal/util"
	"github.com/adrg/xdg"
	"github.com/charlievieth/fastwalk"
)

var (
	paths   []string
	results = providers.QueryData[[]string]{}
)

var terminalApps = make(map[string]struct{})

var (
	Name       = "files"
	NamePretty = "Files"
	config     *Config
)

type Config struct {
	common.Config `koanf:",squash"`
	LaunchPrefix  string `koanf:"launch_prefix" desc:"overrides the default app2unit or uwsm prefix, if set. 'CLEAR' to not prefix." default:""`
}

func init() {
	start := time.Now()

	config = &Config{
		Config: common.Config{
			Icon: "folder",
		},
		LaunchPrefix: "",
	}

	common.LoadConfig(Name, config)

	findTerminalApps()

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

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Search files and folders.")
	fmt.Println()
	util.PrintConfig(Config{}, Name)
}

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)
	results.Lock()
	delete(results.Queries, qid)
	results.Unlock()
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

func Icon() string {
	return config.Icon
}
