package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/abenz1267/elephant/internal/common"
)

const (
	ActionOpen    = "open"
	ActionOpenDir = "opendir"
)

func Activate(qid uint32, identifier, action string) {
	i, err := strconv.Atoi(identifier)
	if err != nil {
		slog.Error(Name, "activate", err)
		return
	}

	switch action {
	case ActionOpen, ActionOpenDir:
		path := paths[i]

		if action == ActionOpenDir {
			path = filepath.Dir(path)
		}

		run := common.WrapWithPrefix(config.LaunchPrefix, fmt.Sprintf("xdg-open '%s'", path))

		if forceTerminalForFile(path) {
			run = wrapWithTerminal(run)
		}

		cmd := exec.Command("sh", "-c", run)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		err := cmd.Start()
		if err != nil {
			slog.Error(Name, "actionopen", err)
		}

		go func() {
			cmd.Wait()
		}()
	default:
		slog.Error(Name, "nosuchaction", action)
	}
}

func wrapWithTerminal(in string) string {
	if terminal == "" {
		return in
	}

	return fmt.Sprintf("%s %s", terminal, in)
}

func forceTerminalForFile(file string) bool {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("xdg-mime query default $(xdg-mime query filetype %s)", file))

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	cmd.Dir = homedir

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return false
	}

	if _, ok := terminalApps[strings.TrimSpace(string(out))]; ok {
		return true
	}

	return false
}
