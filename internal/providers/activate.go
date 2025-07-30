package providers

import "log/slog"

func Activate(sid, qid uint32, provider, identifier, action string) {
	slog.Info("providers", "provider", provider, "identifier", identifier)

	Providers[provider].Activate(qid, identifier, action)

	Cleanup(qid)
}
