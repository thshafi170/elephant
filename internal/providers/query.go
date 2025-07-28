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

var (
	qid atomic.Uint32
	// sessionid => prefix => id
	queries        map[uint32]map[string]uint32
	queryProviders map[uint32][]string
	queryMutex     sync.Mutex
	CleanupChan    chan uint32
)

func init() {
	queries = make(map[uint32]map[string]uint32)
	queryProviders = make(map[uint32][]string)
	CleanupChan = make(chan uint32)

	go func() {
		for qid := range <-CleanupChan {
			Cleanup(qid)
		}
	}()
}

func Query(sid uint32, providers []string, text string, conn net.Conn) {
	start := time.Now()

	queryMutex.Lock()
	if _, ok := queries[sid]; !ok {
		queries[sid] = make(map[string]uint32)
	}
	queryMutex.Unlock()

	var currentQID uint32

	if text != "" {
		lastLength := 1000

		for k, v := range queries[sid] {
			if strings.HasPrefix(text, k) && len(k) < lastLength {
				currentQID = v
				lastLength = len(k)

				queryMutex.Lock()
				delete(queries[sid], k)
				queries[sid][text] = v
				queryMutex.Unlock()
			}
		}

		if currentQID == 0 {
			qid.Add(1)
			currentQID = qid.Load()

			queryMutex.Lock()
			queryProviders[currentQID] = providers
			queries[sid][text] = currentQID
			queryMutex.Unlock()

			slog.Info("providers", "query", "new", "qid", currentQID, "text", text)
		} else {
			slog.Info("providers", "query", "resuming", "qid", currentQID, "text", text)
		}
	} else {
		qid.Add(1)
		currentQID = qid.Load()
		slog.Info("providers", "query", "new", "qid", currentQID, "text", "<empty>")
	}

	conn.Write(fmt.Appendf(nil, "qid;%d\n", currentQID))

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
		conn.Write(fmt.Appendln(nil, "NORESULTS"))
	}

	for _, v := range entries {
		conn.Write(fmt.Appendf(nil, "%d;%s\n", currentQID, Providers[v.Provider].EntryToString(v)))
	}

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
