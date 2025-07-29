package providers

import (
	"sync"
	"sync/atomic"
)

type IterationData[T any] struct {
	ID      atomic.Uint32
	Results map[uint32]T
	sync.Mutex
}

type QueryData[T any] struct {
	Queries map[uint32]*IterationData[T]
	sync.Mutex
}

func (results *QueryData[T]) GetData(qid, iid uint32, data T) (T, bool) {
	results.Lock()
	defer results.Unlock()

	if q, ok := results.Queries[qid]; ok {
		q.Lock()
		defer q.Unlock()
		if iid > q.ID.Load() {
			before := q.ID.Load()
			q.ID.Store(iid)
			results.Queries[qid].Results[iid] = data
			return results.Queries[qid].Results[before], true
		}
	} else {
		results.Queries = make(map[uint32]*IterationData[T])
		results.Queries[qid] = &IterationData[T]{}
		results.Queries[qid].ID.Store(iid)
		results.Queries[qid].Results = make(map[uint32]T)
		results.Queries[qid].Results[iid] = data
		return data, false
	}

	return data, false
}
