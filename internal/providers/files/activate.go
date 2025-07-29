package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
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
	// TODO: find out if it needs to be opened in a terminal, see Walker
	case ActionOpen:

		run := common.WrapWithPrefix(config.LaunchPrefix, fmt.Sprintf("xdg-open '%s'", paths[i]))

		if forceTerminalForFile(paths[i]) {
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
