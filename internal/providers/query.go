package providers

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/abenz1267/elephant/internal/common"
)

var sid atomic.Uint32

var (
	sessions     map[uint32][]string
	sessionMutex sync.Mutex
)

func init() {
	sessions = make(map[uint32][]string)
}

func GetSID() uint32 {
	sid.Add(1)

	return sid.Load()
}

func Query(sid uint32, text string, providers []string, conn net.Conn) {
	start := time.Now()
	sessionMutex.Lock()
	sessions[sid] = providers
	sessionMutex.Unlock()

	slog.Info("providers", "querysession", sid, "query", text)

	var mut sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(providers))

	entries := []common.Entry{}

	for _, v := range providers {
		go func(text string, wg *sync.WaitGroup) {
			defer wg.Done()
			if p, ok := Providers[v]; ok {
				res := p.Query(text)

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

		return 0
	})

	if len(entries) == 0 {
		conn.Write(fmt.Appendln(nil, "NORESULTS"))
	}

	for _, v := range entries {
		conn.Write(fmt.Appendln(nil, v.String()))
	}

	slog.Info("providers", "results", len(entries), "time", time.Since(start))
}

func Activate(sid uint32, provider, identifier, action string) {
	slog.Info("providers", "provider", provider, "identifier", identifier)

	Providers[provider].Activate(sid, identifier, action)

	Cleanup(sid)
}

func Cleanup(sid uint32) {
	slog.Info("providers", "cleanupsession", sid)

	for _, v := range sessions[sid] {
		Providers[v].Cleanup()
	}
}
