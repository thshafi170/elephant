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
	results    = providers.QueryData[map[string]*Symbol]{}
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
		Config:  common.Config{},
		Locale:  "en",
		History: true,
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
	util.PrintConfig(Config{})
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
			h.Save(results.Queries[qid][last].Query, identifier)
		}
	}
}

func Query(qid uint32, iid uint32, query string) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}

	var toFilter map[string]*Symbol

	if query != "" {
		data, ok := results.GetData(query, qid, iid, make(map[string]*Symbol))
		if ok {
			toFilter = data
		} else {
			toFilter = symbols
		}
	} else {
		toFilter = symbols
	}

	for k, v := range toFilter {
		e := &pb.QueryResponse_Item{
			Identifier: k,
			Text:       v.CP,
			Icon:       v.CP,
			Provider:   Name,
			Fuzzyinfo:  &pb.QueryResponse_Item_FuzzyInfo{},
			Type:       pb.QueryResponse_REGULAR,
		}

		if query != "" {
			e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
				Field: "subtext",
			}

			var bestText string
			var bestScore int32
			var bestPos []int32
			var bestStart int32

			for _, m := range v.Searchable {
				score, positions, start := common.FuzzyScore(query, m)

				if score > bestScore {
					bestScore = score
					bestText = m
					bestPos = positions
					bestStart = start
				}
			}

			e.Fuzzyinfo.Positions = bestPos
			e.Fuzzyinfo.Start = bestStart
			e.Score = bestScore
			e.Subtext = bestText

		}

		usage, lastUsed := h.FindUsage(query, e.Identifier)
		e.Score = e.Score + calcUsage(usage, lastUsed)

		if usage != 0 || e.Score > 0 || query == "" {
			if query != "" {
				results.Lock()
				results.Queries[qid][iid].Results[k] = v
				results.Unlock()
			}

			entries = append(entries, e)
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}

func calcUsage(amount int, last time.Time) int32 {
	base := 10

	if amount > 0 {
		today := time.Now()
		duration := today.Sub(last)
		days := int(duration.Hours() / 24)

		if days > 0 {
			base -= days
		}

		res := max(base*amount, 1)

		return int32(res)
	}

	return 0
}
