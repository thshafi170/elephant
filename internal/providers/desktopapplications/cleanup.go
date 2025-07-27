package main

import "log/slog"

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)

	resultsMutex.Lock()
	delete(results, qid)
	resultsMutex.Unlock()
}
