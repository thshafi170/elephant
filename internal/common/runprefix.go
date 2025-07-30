package common

import (
	"log/slog"
	"os/exec"
)

var runPrefix = ""

func InitRunPrefix() {
	app2unit, err := exec.LookPath("app2unit")
	if err == nil && app2unit != "" {
		xdgTerminalExec, err := exec.LookPath("xdg-terminal-exec")
		if err == nil && xdgTerminalExec != "" {
			runPrefix = "app2unit"
			slog.Info("config", "runprefix", runPrefix)
			return
		}
	}

	uwsm, err := exec.LookPath("uwsm")
	if err == nil {
		cmd := exec.Command(uwsm, "check", "is-active")
		err := cmd.Run()
		if err == nil {
			runPrefix = "uwsm app --"
			slog.Info("config", "runprefix", runPrefix)
		}
	}

	if runPrefix == "" {
		slog.Info("config", "runprefix", "<empty>")
	}
}

func LaunchPrefix(override string) string {
	if override == "CLEAR" {
		return ""
	}

	if override != "" {
		return override
	}

	return runPrefix
}
