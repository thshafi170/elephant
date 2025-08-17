package providers

import (
	"sync"
)

type QueryData struct {
	sync.Mutex
	Queries map[uint32]map[uint32]string
}

func (results *QueryData) GetData(query string, qid, iid uint32, exact bool) {
	results.Lock()
	defer results.Unlock()

	if _, ok := results.Queries[qid]; ok {
		if _, ok := results.Queries[qid][iid]; !ok {
			results.Queries[qid][iid] = query
		}

		return
	} else {
		results.Queries = make(map[uint32]map[uint32]string)
		results.Queries[qid] = map[uint32]string{}
		results.Queries[qid][iid] = query

		return
	}
}
