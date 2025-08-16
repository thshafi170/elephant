package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/comm/handlers"
	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/internal/util"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "menues"
	NamePretty = "Menues"
)

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Custom menues.")
	fmt.Println()
	util.PrintConfig(common.MenuConfig{})
}

func Cleanup(qid uint32) {
}

func Activate(qid uint32, identifier, action string, arguments string) {
	var e common.Entry
	var menu common.Menu

	identifier = strings.TrimPrefix(identifier, "keepopen:")
	identifier = strings.TrimPrefix(identifier, "menues:")

	splits := strings.Split(arguments, common.GetElephantConfig().ArgumentDelimiter)
	if len(splits) > 1 {
		arguments = splits[1]
	}

	openmenu := false

	for _, v := range common.Menues {
		if identifier == v.Name {
			menu = v
			openmenu = true
			break
		}

		for _, entry := range v.Entries {
			if identifier == entry.Identifier {
				menu = v
				e = entry
				break
			}
		}
	}

	if openmenu {
		handlers.ProviderUpdated <- fmt.Sprintf("%s:%s", Name, menu.Name)
		return
	}

	run := menu.Action

	if e.Action != "" {
		run = e.Action
	}

	if run == "" {
		return
	}

	pipe := false

	val := e.Value
	if val == "" && len(splits) > 1 {
		val = arguments
	}

	if !strings.Contains(run, "%RESULT%") {
		pipe = true
	} else {
		run = strings.ReplaceAll(run, "%RESULT%", val)
	}

	cmd := exec.Command("sh", "-c", run)

	if pipe && e.Value != "" {
		cmd.Stdin = strings.NewReader(val)
	}

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", err)
	} else {
		go func() {
			cmd.Wait()
		}()
	}
}

func Query(qid uint32, iid uint32, query string, _ bool) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}
	menu := ""

	split := strings.Split(query, ":")
	single := len(split) > 1

	if len(split) > 1 {
		menu = split[0]
		query = split[1]
	}

	for _, v := range common.Menues {
		if menu != "" && v.Name != menu || (!single && !v.GlobalSearch) {
			continue
		}

		icon := v.Icon

		for _, v := range v.Entries {
			if v.Icon != "" {
				icon = v.Icon
			}

			e := &pb.QueryResponse_Item{
				Identifier: v.Identifier,
				Text:       v.Text,
				Subtext:    v.Subtext,
				Provider:   fmt.Sprintf("%s:%s", Name, v.Menu),
				Icon:       icon,
				Type:       pb.QueryResponse_REGULAR,
			}

			if v.Async != "" {
				go func() {
					cmd := exec.Command("sh", "-c", v.Async)
					out, err := cmd.CombinedOutput()

					if err == nil {
						e.Text = strings.TrimSpace(string(out))
					} else {
						e.Text = "%DELETE%"
					}

					providers.AsyncChannels[qid][iid] <- e
				}()
			}

			if query != "" {
				e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
					Field: "text",
				}

				e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start = common.FuzzyScore(query, e.Text)
			}

			if e.Score > 0 || query == "" {
				entries = append(entries, e)
			}
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))

	return entries
}

func Icon() string {
	return ""
}
