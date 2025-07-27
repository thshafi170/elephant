package main

import (
	"log/slog"
	"os"
	"os/exec"

	"github.com/abenz1267/elephant/internal/providers"
)

// uwsm app -- firefox-developer-edition.desktop:open-profile-manager
// app2unit firefox-developer-edition.desktop:new-private-window

var command = ""

func init() {
	app2unit, err := exec.LookPath("app2unit")
	if err == nil && app2unit != "" {
		xdgTerminalExec, err := exec.LookPath("xdg-terminal-exec")
		if err == nil && xdgTerminalExec != "" {
			command = "app2unit"
			slog.Info(Name, "command", command)
			return
		}
	}

	uwsm, err := exec.LookPath("uwsm")
	if err == nil {
		cmd := exec.Command(uwsm, "check", "is-active")
		err := cmd.Run()
		if err == nil {
			command = "uwsm"
			slog.Info(Name, "command", command)
		}
	}

	if command == "" {
		slog.Error(Name, "activation", "no execution command found. Needs app2unit or uwsm.")
		os.Exit(1)
	}
}

func Activate(qid uint32, identifier, action string) {
	cmd := exec.Command(command, identifier)
	if command == "uwsm" {
		cmd = exec.Command("uwsm", "app", "-- ", identifier)
	}

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", identifier, "error", err)
	}

	go func() {
		cmd.Wait()
	}()

	providers.CleanupChan <- qid
}
