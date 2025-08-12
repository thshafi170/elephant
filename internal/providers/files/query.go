package main

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/abenz1267/elephant/internal/comm/pb/pb"
	"github.com/abenz1267/elephant/internal/common"
)

func Query(qid uint32, iid uint32, query string) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}

	var toFilter []string

	if query != "" {
		data, ok := results.GetData(query, qid, iid, []string{})
		if ok {
			toFilter = data
		} else {
			toFilter = paths
		}
	} else {
		toFilter = paths
	}

	slog.Info(Name, "queryingfiles", len(toFilter))

	for k, v := range toFilter {
		common.FuzzyScore(query, v)

		i := strconv.Itoa(k)

		e := &pb.QueryResponse_Item{
			Identifier: i,
			Text:       v,
			Type:       pb.QueryResponse_REGULAR,
			Subtext:    "",
			Provider:   Name,
		}

		if query != "" {
			e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
				Field: "text",
			}

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

	if query != "" {
		results.Lock()
		results.Queries[qid][iid].Done = true
		results.Unlock()
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}
