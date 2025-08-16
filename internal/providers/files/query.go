package main

import (
	"crypto/md5"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

func Query(qid uint32, iid uint32, query string, _ bool, exact bool) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}

	var toFilter []string

	if query != "" {
		data, ok := results.GetData(query, qid, iid, []string{}, exact)
		if ok {
			toFilter = data
		} else {
			toFilter = paths
		}
	} else {
		toFilter = paths
	}

	slog.Info(Name, "queryingfiles", len(toFilter))

	for _, v := range toFilter {
		common.FuzzyScore(query, v, exact)

		md5 := md5.Sum([]byte(v))
		md5str := hex.EncodeToString(md5[:])

		e := &pb.QueryResponse_Item{
			Identifier: md5str,
			Text:       v,
			Type:       pb.QueryResponse_REGULAR,
			Subtext:    "",
			Provider:   Name,
		}

		if query != "" {
			e.Fuzzyinfo = &pb.QueryResponse_Item_FuzzyInfo{
				Field: "text",
			}

			e.Score, e.Fuzzyinfo.Positions, e.Fuzzyinfo.Start = common.FuzzyScore(query, e.Text, exact)
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
