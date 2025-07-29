package main

import (
	"log/slog"
	"os/exec"

	"github.com/abenz1267/elephant/internal/common"
)

func Activate(qid uint32, identifier, action string) {
	cmd := exec.Command("sh", "-c", common.WrapWithPrefix(config.LaunchPrefix, identifier))

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", identifier, "error", err)
	}

	go func() {
		cmd.Wait()
	}()
}
