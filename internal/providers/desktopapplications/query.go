package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
)

var results providers.QueryData[map[string]*DesktopFile]

func init() {
	results = providers.QueryData[map[string]*DesktopFile]{}
}

func Query(qid uint32, iid uint32, query string) []common.Entry {
	start := time.Now()
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	entries := []common.Entry{}

	var toFilter map[string]*DesktopFile

	data, ok := results.GetData(qid, iid, make(map[string]*DesktopFile))
	if ok {
		toFilter = data
	} else {
		toFilter = files
	}

	slog.Info(Name, "queryingfiles", len(toFilter))

	for k, v := range toFilter {
		if len(v.NotShowIn) != 0 && slices.Contains(v.NotShowIn, desktop) || len(v.OnlyShowIn) != 0 && !slices.Contains(v.OnlyShowIn, desktop) || v.Hidden || v.NoDisplay {
			continue
		}

		// add generic entry
		if config.ShowActions && config.ShowGeneric || !config.ShowActions || (config.ShowActions && len(v.Actions) == 0) || query == "" {
			e := common.Entry{
				Identifier: k,
				Text:       v.Name,
				SubText:    v.GenericName,
				Icon:       v.Icon,
				Provider:   Name,
			}

			var match string
			var ok bool

			if query != "" {
				e.Fuzzy = &common.FuzzyMatchInfo{
					Field: "text",
				}

				match, e.Score, e.Fuzzy.Pos, e.Fuzzy.Start, ok = calcScore(query, &v.Data)

				if ok && match != e.Text {
					e.SubText = match
					e.Fuzzy.Field = "subtext"
				}
			}

			if e.Score > 0 || query == "" {
				entries = append(entries, e)

				results.Queries[qid].Results[iid][k] = v
			}
		}

		// add actions
		if (query == "" && config.ShowActionsWithoutQuery) || (query != "" && config.ShowActions) {
			for _, a := range v.Actions {

				e := common.Entry{
					Identifier: fmt.Sprintf("%s:%s", k, a.Action),
					Text:       a.Name,
					SubText:    v.Name,
					Icon:       a.Icon,
					Provider:   Name,
				}

				var match string
				var ok bool

				if query != "" {
					e.Fuzzy = &common.FuzzyMatchInfo{
						Field: "text",
					}

					match, e.Score, e.Fuzzy.Pos, e.Fuzzy.Start, ok = calcScore(query, &a)

					if ok && match != e.Text {
						e.SubText = match
						e.Fuzzy.Field = "subtext"
					}
				}

				if e.Score > 0 || query == "" {
					entries = append(entries, e)

					results.Queries[qid].Results[iid][k] = v
				}
			}
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))

	return entries
}

func calcScore(q string, d *Data) (string, int, *[]int, int, bool) {
	var scoreRes int
	var posRes *[]int
	var startRes int
	var match string
	var modifier int

	for k, v := range []string{d.Name, d.Parent, d.GenericName, strings.Join(d.Keywords, ","), d.Comment} {
		score, pos, start := common.FuzzyScore(q, v)

		if score > scoreRes {
			scoreRes = score
			posRes = pos
			startRes = start
			match = v
			modifier = k
		}
	}

	if scoreRes == 0 {
		return "", 0, nil, 0, false
	}

	scoreRes = max(scoreRes-min(modifier*10, 50)-startRes, 10)

	return match, scoreRes, posRes, startRes, true
}

func EntryToString(e common.Entry) string {
	var start int
	var field string

	positions := []string{}

	if e.Fuzzy != nil {
		if e.Fuzzy.Pos != nil {
			for _, num := range *e.Fuzzy.Pos {
				positions = append(positions, strconv.Itoa(num))
			}
		}

		start = e.Fuzzy.Start
		field = e.Fuzzy.Field
	}

	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;%d;%s", e.Provider, e.Identifier, e.Text, e.SubText, e.Icon, strings.Join(positions, ","), start, field)
}
