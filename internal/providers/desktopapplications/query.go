package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var results = providers.QueryData{}

func Query(qid uint32, iid uint32, query string, _ bool, exact bool) []*pb.QueryResponse_Item {
	start := time.Now()
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	entries := []*pb.QueryResponse_Item{}

	isSub := qid >= 100_000_000

	if !isSub && query != "" {
		results.GetData(query, qid, iid, exact)
	}

	for k, v := range files {
		if len(v.NotShowIn) != 0 && slices.Contains(v.NotShowIn, desktop) || len(v.OnlyShowIn) != 0 && !slices.Contains(v.OnlyShowIn, desktop) || v.Hidden || v.NoDisplay {
			continue
		}

		// check generic
		e := &pb.QueryResponse_Item{
			Identifier: k,
			Text:       v.Name,
			Type:       pb.QueryResponse_REGULAR,
			Subtext:    v.GenericName,
			Icon:       v.Icon,
			Provider:   Name,
		}

		var match string
		var ok bool

		if query != "" {
			e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
				Field: "text",
			}

			match, e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start, ok = calcScore(query, &v.Data, exact)

			if ok && match != e.Text {
				e.Subtext = match
				e.Fuzzyinfo.Field = "subtext"
			}
		}

		var usageScore int32
		if config.History && (e.Score > 0 || query == "") {
			usageScore = h.CalcUsageScore(query, e.Identifier)
			e.Score = e.Score + usageScore
		}

		if usageScore != 0 || config.ShowActions && config.ShowGeneric || !config.ShowActions || (config.ShowActions && len(v.Actions) == 0) || query == "" {
			if e.Score >= config.MinScore || query == "" {
				entries = append(entries, e)
			}
		}

		// check actions
		for _, a := range v.Actions {
			e := &pb.QueryResponse_Item{
				Identifier: fmt.Sprintf("%s:%s", k, a.Action),
				Text:       a.Name,
				Type:       pb.QueryResponse_REGULAR,
				Subtext:    v.Name,
				Icon:       a.Icon,
				Provider:   Name,
			}

			var match string
			var ok bool

			if query != "" {
				e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
					Field: "text",
				}

				match, e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start, ok = calcScore(query, &a, exact)

				if ok && match != e.Text {
					e.Subtext = match
					e.Fuzzyinfo.Field = "subtext"
				}
			}

			var usageScore int32
			if config.History && (e.Score > 0 || query == "") {
				usageScore = h.CalcUsageScore(query, e.Identifier)
				e.Score = e.Score + usageScore
			}

			if (query == "" && config.ShowActionsWithoutQuery) || (query != "" && config.ShowActions) || usageScore != 0 {
				if e.Score >= config.MinScore || query == "" {
					entries = append(entries, e)
				}
			}
		}
	}

	if !isSub {
		slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	}

	return entries
}

func calcScore(q string, d *Data, exact bool) (string, int32, []int32, int32, bool) {
	var scoreRes int32
	var posRes []int32
	var startRes int32
	var match string
	var modifier int32

	for k, v := range []string{d.Name, d.Parent, d.GenericName, strings.Join(d.Keywords, ","), d.Comment} {
		score, pos, start := common.FuzzyScore(q, v, exact)

		if score > scoreRes {
			scoreRes = score
			posRes = pos
			startRes = start
			match = v
			modifier = int32(k)
		}
	}

	if scoreRes == 0 {
		return "", 0, nil, 0, false
	}

	scoreRes = max(scoreRes-min(modifier*10, 50)-startRes, 10)

	return match, scoreRes, posRes, startRes, true
}
