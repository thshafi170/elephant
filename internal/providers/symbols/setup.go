// Package symbols provides symbols/emojis.
package main

import (
	"fmt"
	"log"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/common/history"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/internal/util"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "symbols"
	NamePretty = "Symbols/Emojis"
	h          = history.Load(Name)
	results    = providers.QueryData{}
)

type Config struct {
	common.Config `koanf:",squash"`
	Locale        string `koanf:"locale" desc:"locale to use for symbols" default:"en"`
	History       bool   `koanf:"history" desc:"make use of history for sorting" default:"false"`
}

var config *Config

func init() {
	start := time.Now()

	config = &Config{
		Config: common.Config{
			Icon:     "face-smile",
			MinScore: 50,
		},
		Locale:  "en",
		History: false,
	}

	common.LoadConfig(Name, config)

	parse()

	slog.Info(Name, "symbols/emojis", len(symbols), "time", time.Since(start))
}

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Find symbols and emojis.")
	fmt.Println()
	fmt.Println("Possible locales:")

	entries, err := files.ReadDir("data")
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range entries {
		fmt.Printf("%s,", strings.TrimSuffix(filepath.Base(v.Name()), ".xml"))
	}

	fmt.Println()
	fmt.Println()
	util.PrintConfig(Config{}, Name)
}

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)
	results.Lock()
	delete(results.Queries, qid)
	results.Unlock()
}

func Activate(qid uint32, identifier, action string, arguments string) {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(symbols[identifier].CP)

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", err)
	} else {
		go func() {
			cmd.Wait()
		}()
	}

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

func Query(qid uint32, iid uint32, query string, _ bool, exact bool) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}

	if query != "" {
		results.GetData(query, qid, iid, exact)
	}

	for k, v := range symbols {
		field := "subtext"
		var positions []int32
		var fs int32
		var score int32
		var text string

		if query != "" {
			var bestText string
			var bestScore int32
			var bestPos []int32
			var bestStart int32

			for _, m := range v.Searchable {
				score, positions, start := common.FuzzyScore(query, m, exact)

				if score > bestScore {
					bestScore = score
					bestText = m
					bestPos = positions
					bestStart = start
				}
			}

			positions = bestPos
			fs = bestStart
			score = bestScore
			text = bestText
		} else {
			text = v.Searchable[len(v.Searchable)-1]
		}

		var usageScore int32
		if config.History && (score > config.MinScore || query == "") {
			usageScore = h.CalcUsageScore(query, k)
			score = score + usageScore
		}

		if usageScore != 0 || score > config.MinScore || query == "" {
			entries = append(entries, &pb.QueryResponse_Item{
				Identifier: k,
				Score:      score,
				Text:       text,
				Icon:       v.CP,
				Provider:   Name,
				Fuzzyinfo: &pb.QueryResponse_Item_FuzzyInfo{
					Start:     fs,
					Field:     field,
					Positions: positions,
				},
				Type: pb.QueryResponse_REGULAR,
			})
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}

func Icon() string {
	return config.Icon
}
