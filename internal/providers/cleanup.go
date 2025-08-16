package providers

import (
	"log/slog"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type timestamps struct {
	Data map[uint32]time.Time
	sync.Mutex
}

// this is used to auto-clean query data
var Timestampedqueries *timestamps

func init() {
	Timestampedqueries = &timestamps{
		Data: make(map[uint32]time.Time),
	}

	go func() {
		for {
			time.Sleep(1 * time.Minute)

			now := time.Now()

			for k, v := range Timestampedqueries.Data {
				if now.Sub(v).Seconds() > 60 {
					Cleanup(k)

					Timestampedqueries.Lock()
					delete(Timestampedqueries.Data, k)
					Timestampedqueries.Unlock()
				}
			}

			runtime.GC()
			debug.FreeOSMemory()
		}
	}()
}

func Cleanup(qid uint32) {
	slog.Info("providers", "cleanup", qid)

	for _, v := range AsyncChannels[qid] {
		close(v)
	}

	delete(AsyncChannels, qid)

	for _, v := range QueryProviders[qid] {
		if p, ok := Providers[v]; ok {
			p.Cleanup(qid)
		}
	}
}
