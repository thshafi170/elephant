package providers

import (
	"strings"
	"sync"
)

type Iteration[T any] struct {
	Query   string
	Done    bool
	Results T
}

type QueryData[T any] struct {
	sync.Mutex
	Queries map[uint32]map[uint32]*Iteration[T]
}

func (results *QueryData[T]) GetData(query string, qid, iid uint32, data T) (T, bool) {
	results.Lock()
	defer results.Unlock()

	if q, ok := results.Queries[qid]; ok {
		if _, ok := results.Queries[qid][iid]; !ok {
			results.Queries[qid][iid] = &Iteration[T]{Results: data, Query: query}
		}

		var longestid uint32
		var longest int

		for i, v := range q {
			if strings.HasPrefix(query, v.Query) && v.Done && len(v.Query) > longest {
				longestid = i
				longest = len(v.Query)
			}
		}

		if longestid != 0 {
			return q[longestid].Results, true
		}

		return data, false
	} else {
		results.Queries = make(map[uint32]map[uint32]*Iteration[T])
		results.Queries[qid] = map[uint32]*Iteration[T]{}
		results.Queries[qid][iid] = &Iteration[T]{Results: data, Query: query}

		return data, false
	}
}
