package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/abenz1267/elephant/internal/common"
)

func Activate(qid uint32, identifier, action string, arguments string) {
	toRun := ""
	prefix := common.LaunchPrefix(config.LaunchPrefix)

	splits := strings.Split(arguments, common.GetElephantConfig().ArgumentDelimiter)
	if len(splits) > 1 {
		arguments = splits[1]
	} else {
		arguments = ""
	}

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

	toRun = strings.TrimSpace(fmt.Sprintf("%s %s", toRun, arguments))

	cmd := exec.Command("sh", "-c", toRun)

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", identifier, "error", err)
	}

	go func() {
		cmd.Wait()
	}()

	if config.History {
		var last uint32

		for k := range results.Queries[qid] {
			if k > last {
				last = k
			}
		}

		if last != 0 {
			h.Save(results.Queries[qid][last], identifier)
		} else {
			h.Save("", identifier)
		}
	}
}
