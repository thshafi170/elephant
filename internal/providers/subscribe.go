package providers

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"sync/atomic"
	"time"

	"github.com/abenz1267/elephant/internal/common"
)

var (
	sid             atomic.Uint32
	subs            map[uint32]*sub
	subResults      chan *sub
	ProviderUpdated chan string
)

type sub struct {
	sid      uint32
	interval int
	provider string
	query    string
	results  []common.Entry
	conn     net.Conn
}

func init() {
	sid.Store(100_000_000)
	subs = make(map[uint32]*sub)
	subResults = make(chan *sub)
	ProviderUpdated = make(chan string)

	// handle queried subs
	go func() {
		for res := range subResults {
			res.conn.Write(fmt.Appendf(nil, "sid;%d;changed\n", res.sid))
		}
	}()

	// handle general realtime subs
	go func() {
		for p := range ProviderUpdated {
			for _, v := range subs {
				if v.provider == p && v.interval == 0 && v.query == "" {
					v.conn.Write(fmt.Appendf(nil, "sid;%d;changed\n", v.sid))
				}
			}
		}
	}()
}

func Subscribe(interval int, provider, query string, conn net.Conn) {
	sid.Add(1)

	sub := &sub{
		sid:      sid.Load(),
		interval: interval,
		provider: provider,
		query:    query,
		conn:     conn,
		results:  []common.Entry{},
	}

	subs[sub.sid] = sub

	if interval != 0 {
		go watch(sub)
	}

	conn.Write(fmt.Appendf(nil, "sid;%d;subscribed\n", sub.sid))
}

func Unsubscribe(sid uint32) {
	delete(subs, sid)
	slog.Info("providers", "unsubscribe", sid)
}

func watch(s *sub) {
	p := Providers[s.provider]

	for {
		time.Sleep(time.Duration(s.interval) * time.Millisecond)

		if _, ok := subs[s.sid]; !ok {
			return
		}

		res := p.Query(s.sid, s.sid, s.query)

		slices.SortFunc(res, sortEntries)

		if len(s.results) != 0 {
			if len(res) != len(s.results) {
				s.results = res
				subResults <- s
				continue
			}

			for k, v := range res {
				if p.EntryToString(v) != p.EntryToString(s.results[k]) {
					s.results = res
					subResults <- s
					break
				}
			}
		} else {
			s.results = res
		}
	}
}
