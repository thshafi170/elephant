package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/abenz1267/elephant/internal/comm/handlers"
	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/util"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "websearch"
	NamePretty = "Websearch"
	config     *Config
)

const (
	ActionCopy   = "copy"
	ActionSave   = "save"
	ActionDelete = "delete"
)

type Config struct {
	common.Config           `koanf:",squash"`
	Entries                 []Entry `koanf:"entries" desc:"entries" default:""`
	MaxGlobalItemsToDisplay int     `koanf:"max_global_items_to_display" desc:"will only show the global websearch entry if there are at most X results." default:"1"`
}

type Entry struct {
	Name    string `koanf:"name" desc:"name of the entry" default:""`
	Default bool   `koanf:"default" desc:"entry to display when querying multiple providers" default:""`
	Prefix  string `koanf:"prefix" desc:"prefix to actively trigger this entry" default:""`
	URL     string `koanf:"url" desc:"url, example: 'https://www.google.com/search?q=%TERM%'" default:""`
	Icon    string `koanf:"icon" desc:"icon to display, fallsback to global" default:""`
}

var prefixes = make(map[string]int)

func init() {
	config = &Config{
		Config: common.Config{
			Icon: "applications-internet",
		},
		MaxGlobalItemsToDisplay: 1,
	}

	common.LoadConfig(Name, config)
	handlers.MaxGlobalItemsToDisplayWebsearch = config.MaxGlobalItemsToDisplay

	for k, v := range config.Entries {
		if v.Prefix != "" {
			prefixes[v.Prefix] = k
			handlers.WebsearchPrefixes[v.Prefix] = v.Name
		}
	}
}

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Websearch: search the web with custom defined searches")
	fmt.Println()
	util.PrintConfig(Config{}, Name)
}

func Cleanup(qid uint32) {
}

func Activate(qid uint32, identifier, action string, query string) {
	i, _ := strconv.Atoi(identifier)

	for k := range prefixes {
		if after, ok := strings.CutPrefix(query, k); ok {
			query = after
			break
		}
	}

	url := strings.ReplaceAll(config.Entries[i].URL, "%TERM%", url.QueryEscape(query))

	prefix := common.LaunchPrefix("")

	cmd := exec.Command("sh", "-c", strings.TrimSpace(fmt.Sprintf("%s xdg-open '%s'", prefix, url)))

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
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

func Query(qid uint32, iid uint32, query string, single bool, _ bool) []*pb.QueryResponse_Item {
	entries := []*pb.QueryResponse_Item{}

	prefix := ""

	for k := range prefixes {
		if strings.HasPrefix(query, k) {
			prefix = k
			break
		}
	}

	if single {
		for k, v := range config.Entries {
			icon := v.Icon
			if icon == "" {
				icon = config.Icon
			}

			e := &pb.QueryResponse_Item{
				Identifier: strconv.Itoa(k),
				Text:       v.Name,
				Subtext:    "",
				Icon:       icon,
				Provider:   Name,
				Score:      int32(100 - k),
				Type:       0,
			}

			entries = append(entries, e)
		}
	} else {
		for k, v := range config.Entries {
			if v.Default || v.Prefix == prefix {
				icon := v.Icon
				if icon == "" {
					icon = config.Icon
				}

				e := &pb.QueryResponse_Item{
					Identifier: strconv.Itoa(k),
					Text:       v.Name,
					Subtext:    "",
					Icon:       icon,
					Provider:   Name,
					Score:      int32(100 - k),
					Type:       0,
				}

				entries = append(entries, e)
			}
		}
	}

	return entries
}

func Icon() string {
	return config.Icon
}
