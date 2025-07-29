package main

import "log/slog"

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)

	results.Lock()
	delete(results.Queries, qid)
	results.Unlock()
}
