package providers

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/abenz1267/elephant/internal/common"
)

type QueryData struct {
	Query     string
	Iteration atomic.Uint32
	sync.Mutex
}

var (
	qid atomic.Uint32
	// sessionid => prefix => id
	queries        map[uint32]map[uint32]*QueryData
	queryProviders map[uint32][]string
	queryMutex     sync.Mutex
)

func init() {
	queries = make(map[uint32]map[uint32]*QueryData)
	queryProviders = make(map[uint32][]string)
}

func Query(sid uint32, providers []string, text string, conn net.Conn) {
	start := time.Now()

	queryMutex.Lock()
	if _, ok := queries[sid]; !ok {
		queries[sid] = make(map[uint32]*QueryData)
	}
	queryMutex.Unlock()

	var currentQID uint32
	var currentIteration uint32

	if text != "" {
		lastLength := 1000

		for k, v := range queries[sid] {
			if strings.HasPrefix(text, v.Query) && len(v.Query) < lastLength {
				currentQID = k
				lastLength = len(v.Query)
				v.Iteration.Add(1)
				currentIteration = v.Iteration.Load()
			}
		}

		if currentQID == 0 {
			qid.Add(1)
			currentQID = qid.Load()

			queryMutex.Lock()
			queryProviders[currentQID] = providers
			data := &QueryData{
				Query: text,
			}
			data.Iteration.Add(1)
			currentIteration = data.Iteration.Load()
			queries[sid][currentQID] = data
			queryMutex.Unlock()

			slog.Info("providers", "query", "new", "qid", currentQID, "iid", currentIteration, "text", text)
		} else {
			slog.Info("providers", "query", "resuming", "qid", currentQID, "iid", currentIteration, "text", text)
		}
	} else {
		qid.Add(1)
		currentQID = qid.Load()
		currentIteration = 1
		slog.Info("providers", "query", "new", "qid", currentQID, "iid", currentIteration, "text", "<empty>")
	}

	conn.Write(fmt.Appendf(nil, "qid;%d;%d\n", currentQID, currentIteration))

	var mut sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(providers))

	entries := []common.Entry{}

	for _, v := range providers {
		go func(text string, wg *sync.WaitGroup) {
			defer wg.Done()
			if p, ok := Providers[v]; ok {
				res := p.Query(currentQID, text)

				mut.Lock()
				entries = append(entries, res...)
				mut.Unlock()
			}
		}(text, &wg)
	}

	wg.Wait()

	slices.SortFunc(entries, func(a common.Entry, b common.Entry) int {
		if a.Score > b.Score {
			return -1
		}

		if b.Score > a.Score {
			return 1
		}

		return strings.Compare(a.Text, b.Text)
	})

	if len(entries) == 0 {
		conn.Write(fmt.Appendf(nil, "qid;%d;noresults", currentQID))
	}

	for _, v := range entries {
		if text != "" && currentIteration != queries[sid][currentQID].Iteration.Load() {
			slog.Info("providers", "results", "aborting", "qid", currentQID, "iid", currentIteration)
			return
		}

		conn.Write(fmt.Appendf(nil, "%d;%d;%s\n", currentQID, currentIteration, Providers[v.Provider].EntryToString(v)))
	}

	conn.Write(fmt.Appendf(nil, "qid;%d;done\n", currentQID))

	slog.Info("providers", "results", len(entries), "time", time.Since(start))
}

func Activate(sid, qid uint32, provider, identifier, action string) {
	slog.Info("providers", "provider", provider, "identifier", identifier)

	Providers[provider].Activate(qid, identifier, action)

	Cleanup(qid)
}

func Cleanup(qid uint32) {
	slog.Info("providers", "cleanup", qid)

	for _, v := range queryProviders[qid] {
		Providers[v].Cleanup(qid)
	}
}
