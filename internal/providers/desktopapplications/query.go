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

var results providers.QueryData[map[string]*DesktopFile]

func init() {
	results = providers.QueryData[map[string]*DesktopFile]{}
}

func Query(qid uint32, iid uint32, query string) []*pb.QueryResponse_Item {
	start := time.Now()
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	entries := []*pb.QueryResponse_Item{}

	var toFilter map[string]*DesktopFile

	isSub := qid >= 100_000_000

	if !isSub && query != "" {
		data, ok := results.GetData(query, qid, iid, make(map[string]*DesktopFile))
		if ok {
			toFilter = data
		} else {
			toFilter = files
		}
	} else {
		toFilter = files
	}

	slog.Info(Name, "queryingfiles", len(toFilter))

	for k, v := range toFilter {
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

			match, e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start, ok = calcScore(query, &v.Data)

			if ok && match != e.Text {
				e.Subtext = match
				e.Fuzzyinfo.Field = "subtext"
			}
		}

		usage, lastUsed := h.FindUsage(query, e.Identifier)
		e.Score = e.Score + calcUsage(usage, lastUsed)

		if usage != 0 || config.ShowActions && config.ShowGeneric || !config.ShowActions || (config.ShowActions && len(v.Actions) == 0) || query == "" {
			if e.Score > 0 || query == "" {
				entries = append(entries, e)

				if !isSub && query != "" {
					results.Lock()
					results.Queries[qid][iid].Results[k] = v
					results.Unlock()
				}
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

				match, e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start, ok = calcScore(query, &a)

				if ok && match != e.Text {
					e.Subtext = match
					e.Fuzzyinfo.Field = "subtext"
				}
			}

			usage, lastUsed := h.FindUsage(query, e.Identifier)
			e.Score = e.Score + calcUsage(usage, lastUsed)

			if (query == "" && config.ShowActionsWithoutQuery) || (query != "" && config.ShowActions) || usage != 0 {
				if e.Score > 0 || query == "" {
					entries = append(entries, e)

					if !isSub && query != "" {
						results.Lock()
						results.Queries[qid][iid].Results[k] = v
						results.Unlock()
					}
				}
			}
		}
	}

	if !isSub && query != "" {
		results.Lock()
		results.Queries[qid][iid].Done = true
		results.Unlock()
	}

	if !isSub {
		slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	}

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

		res := base * amount

		if res < 1 {
			res = 1
		}

		return int32(res)
	}

	return 0
}

func calcScore(q string, d *Data) (string, int32, []int32, int32, bool) {
	var scoreRes int32
	var posRes []int32
	var startRes int32
	var match string
	var modifier int32

	for k, v := range []string{d.Name, d.Parent, d.GenericName, strings.Join(d.Keywords, ","), d.Comment} {
		score, pos, start := common.FuzzyScore(q, v)

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
