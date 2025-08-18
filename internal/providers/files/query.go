package main

import (
	"crypto/md5"
	"encoding/hex"
	"log/slog"
	"strings"
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

	for k, v := range paths {
		md5 := md5.Sum([]byte(k))
		md5str := hex.EncodeToString(md5[:])

		e := &pb.QueryResponse_Item{
			Identifier: md5str,
			Text:       k,
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

		if !strings.HasSuffix(k, "/") && query == "" {
			e.Score = e.Score + calcScore(v, start)
		}

		if e.Score > 0 || query == "" {
			entries = append(entries, e)
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}

func calcScore(v time.Time, now time.Time) int32 {
	diff := now.Sub(v)

	res := 3600 - diff.Seconds()

	if res < 0 {
		res = 0
	}

	return int32(res)
}
