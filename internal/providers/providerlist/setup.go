package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "providerlist"
	NamePretty = "Providerlist"
)

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("List installed providers")
	fmt.Println()
}

func Cleanup(qid uint32) {
}

func Activate(qid uint32, identifier, action string, arguments string) {
}

func Query(qid uint32, iid uint32, query string, single bool) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}

	for _, v := range providers.Providers {
		if *v.Name == Name {
			continue
		}

		e := &pb.QueryResponse_Item{
			Identifier: *v.Name,
			Text:       *v.NamePretty,
			Provider:   Name,
			Type:       pb.QueryResponse_REGULAR,
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

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))

	slices.SortFunc(entries, func(a, b *pb.QueryResponse_Item) int {
		if a.Score > b.Score {
			return 1
		}

		if a.Score < b.Score {
			return -1
		}

		return strings.Compare(a.Text, b.Text)
	})

	return entries
}
