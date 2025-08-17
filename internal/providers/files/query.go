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

	if query != "" {
		results.GetData(query, qid, iid, exact)
	}

	for _, v := range paths {
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
			entries = append(entries, e)
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}
