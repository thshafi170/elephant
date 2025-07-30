package providers

import (
	"log/slog"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type timestamps struct {
	data map[uint32]time.Time
	sync.Mutex
}

func (t *timestamps) remove(qid uint32) {
	t.Lock()
	delete(t.data, qid)
	t.Unlock()
}

var timestampedqueries *timestamps

func init() {
	timestampedqueries = &timestamps{
		data: make(map[uint32]time.Time),
	}

	go func() {
		for {
			time.Sleep(1 * time.Minute)

			now := time.Now()

			for k, v := range timestampedqueries.data {
				if now.Sub(v).Seconds() > 60 {
					Cleanup(k)

					timestampedqueries.Lock()
					delete(timestampedqueries.data, k)
					timestampedqueries.Unlock()
				}
			}

			runtime.GC()
			debug.FreeOSMemory()
		}
	}()
}

func Cleanup(qid uint32) {
	slog.Info("providers", "cleanup", qid)

	for _, v := range queryProviders[qid] {
		Providers[v].Cleanup(qid)
	}
}
