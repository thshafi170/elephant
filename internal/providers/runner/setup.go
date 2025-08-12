// Package runner provides access to binaries in $PATH.
package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/abenz1267/elephant/internal/comm/pb/pb"
	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
)

var (
	Name       = "runner"
	NamePretty = "Runner"
)

var bins = []string{}

func Load() {
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
	i, _ := strconv.Atoi(identifier)
	cmd := exec.Command("sh", "-c", strings.TrimSpace(fmt.Sprintf("%s %s", bins[i], arguments)))

	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "activate", err)
	} else {
		go func() {
			cmd.Wait()
		}()
	}
}

var results providers.QueryData[[]string]

func init() {
	results = providers.QueryData[[]string]{}
}

func Query(qid uint32, iid uint32, query string) []*pb.QueryResponse_Item {
	entries := []*pb.QueryResponse_Item{}

	var toFilter []string

	if query != "" {
		data, ok := results.GetData(query, qid, iid, []string{})
		if ok {
			toFilter = data
		} else {
			toFilter = bins
		}
	} else {
		toFilter = bins
	}

	for k, v := range toFilter {
		e := &pb.QueryResponse_Item{
			Identifier: strconv.Itoa(k),
			Text:       v,
			Provider:   Name,
			Score:      0,
			Fuzzyinfo:  &pb.QueryResponse_Item_FuzzyInfo{},
			Type:       pb.QueryResponse_REGULAR,
		}

		if query != "" {
			e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start = common.FuzzyScore(query, e.Text)
		}

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
