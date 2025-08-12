// Package runner provides access to binaries in $PATH.
package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/common/history"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "runner"
	NamePretty = "Runner"
	results    = providers.QueryData[[]Item]{}
)

type Config struct {
	common.Config `koanf:",squash"`
	History       bool `koanf:"history" desc:"make use of history for sorting" default:"false"`
}

var (
	config *Config
	items  = []Item{}
	h      = history.Load(Name)
)

type Item struct {
	Identifier string
	Bin        string
}

func init() {
	start := time.Now()

	config = &Config{
		Config:  common.Config{},
		History: true,
	}

	common.LoadConfig(Name, config)

	bins := []string{}

	for p := range strings.SplitSeq(os.Getenv("PATH"), ":") {
		filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if d != nil && d.IsDir() {
				return nil
			}

			info, serr := os.Stat(path)
			if info == nil || serr != nil {
				return nil
			}

			if info.Mode()&0111 != 0 {
				bins = append(bins, filepath.Base(path))
			}

			return nil
		})
	}

	bins = slices.Compact(bins)

	for _, v := range bins {
		md5 := md5.Sum([]byte(v))
		md5str := hex.EncodeToString(md5[:])

		items = append(items, Item{
			Identifier: md5str,
			Bin:        v,
		})
	}

	slog.Info(Name, "executables", len(items), "time", time.Since(start))
}

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Run everything in your $PATH!")
	fmt.Println()
}

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)
	results.Lock()
	delete(results.Queries, qid)
	results.Unlock()
}

func Activate(qid uint32, identifier, action string, arguments string) {
	bin := ""

	for _, v := range items {
		if v.Identifier == identifier {
			bin = v.Bin
			break
		}
	}

	cmd := exec.Command("sh", "-c", strings.TrimSpace(fmt.Sprintf("%s %s", bin, arguments)))

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
	entries := []*pb.QueryResponse_Item{}

	var toFilter []Item

	if query != "" {
		data, ok := results.GetData(query, qid, iid, []Item{})
		if ok {
			toFilter = data
		} else {
			toFilter = items
		}
	} else {
		toFilter = items
	}

	for _, v := range toFilter {
		e := &pb.QueryResponse_Item{
			Identifier: v.Identifier,
			Text:       v.Bin,
			Provider:   Name,
			Score:      0,
			Fuzzyinfo:  &pb.QueryResponse_Item_FuzzyInfo{},
			Type:       pb.QueryResponse_REGULAR,
		}

		if query != "" {
			e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start = common.FuzzyScore(query, e.Text)
		}

		usageScore := h.CalcUsageScore(query, e.Identifier)
		e.Score = e.Score + usageScore

		if e.Score > 0 || query == "" {
			if query != "" {
				results.Lock()
				results.Queries[qid][iid].Results = append(results.Queries[qid][iid].Results, v)
				results.Unlock()
			}

			entries = append(entries, e)
		}
	}

	return entries
}
