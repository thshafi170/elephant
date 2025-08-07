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

		var match string
		var ok bool

		if query != "" {
			e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
				Field: "text",
			}

			match, e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start, ok = calcScore(query, v)

			if ok && match != e.Text {
				e.Subtext = match
				e.Fuzzyinfo.Field = "text"
			}
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

func calcScore(q string, d string) (string, int32, []int32, int32, bool) {
	var scoreRes int32
	var posRes []int32
	var startRes int32
	var match string
	var modifier int32

	score, pos, start := common.FuzzyScore(q, d)

	if score > scoreRes {
		scoreRes = score
		posRes = pos
		startRes = start
		match = d
		modifier = 0
	}

	if scoreRes == 0 {
		return "", 0, nil, 0, false
	}

	scoreRes = max(scoreRes-min(modifier*10, 50)-startRes, 10)

	return match, scoreRes, posRes, startRes, true
}
