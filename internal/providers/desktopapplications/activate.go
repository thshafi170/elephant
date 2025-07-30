package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/abenz1267/elephant/internal/common"
)

func Activate(qid uint32, identifier, action string) {
	toRun := ""
	prefix := common.LaunchPrefix(config.LaunchPrefix)

	if prefix == "" {
		parts := strings.Split(identifier, ":")

		if len(parts) == 2 {
			for _, v := range files[parts[0]].Actions {
				if v.Action == parts[1] {
					toRun = v.Exec
					break
				}
			}
		} else {
			toRun = files[parts[0]].Exec
		}
	} else {
		toRun = fmt.Sprintf("%s %s", prefix, identifier)
	}

	cmd := exec.Command("sh", "-c", toRun)
	fmt.Println(cmd.String())

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", identifier, "error", err)
	}

	go func() {
		cmd.Wait()
	}()
}
