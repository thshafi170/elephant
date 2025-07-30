package main

import (
	"log/slog"
	"os/exec"
	"strings"

	"github.com/abenz1267/elephant/internal/common"
)

func Activate(qid uint32, identifier, action string) {
	if config.LaunchPrefix == "" {
		identifier = strings.Split(identifier, ":")[0]
	}

	cmd := exec.Command("sh", "-c", common.WrapWithPrefix(config.LaunchPrefix, identifier))

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", identifier, "error", err)
	}

	go func() {
		cmd.Wait()
	}()
}
