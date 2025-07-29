package main

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/abenz1267/elephant/internal/common"
)

func Query(qid uint32, iid uint32, text string) []common.Entry {
	start := time.Now()
	entries := []common.Entry{}

	var toFilter []string

	data, ok := results.GetData(qid, iid, []string{})
	if ok {
		toFilter = data
	} else {
		toFilter = paths
	}

	slog.Info(Name, "queryingfiles", len(toFilter))

	for k, v := range toFilter {
		common.FuzzyScore(text, v)

		i := strconv.Itoa(k)

		e := common.Entry{
			Identifier: i,
			Text:       v,
			SubText:    "",
			Provider:   Name,
		}

		var match string
		var ok bool

		if text != "" {
			e.Fuzzy = &common.FuzzyMatchInfo{
				Field: "text",
			}

			match, e.Score, e.Fuzzy.Pos, e.Fuzzy.Start, ok = calcScore(text, v)

			if ok && match != e.Text {
				e.SubText = match
				e.Fuzzy.Field = "text"
			}
		}

		if e.Score > 0 || text == "" {
			results.Queries[qid].Lock()
			results.Queries[qid].Results[iid] = append(results.Queries[qid].Results[iid], v)
			results.Queries[qid].Unlock()

			entries = append(entries, e)
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}

func calcScore(q string, d string) (string, int, *[]int, int, bool) {
	var scoreRes int
	var posRes *[]int
	var startRes int
	var match string
	var modifier int

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
