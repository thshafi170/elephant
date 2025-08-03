// Package history provides functions to save and load history in a streamlined way.
package history

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/abenz1267/elephant/internal/common"
)

type HistoryData struct {
	LastUsed time.Time
	Amount   int
}

// TODO: this is global for every history ... should not be the case. Just a crutch because of gob encoding.
var mut sync.Mutex

type History struct {
	Provider string
	Data     map[string]map[string]*HistoryData
}

func (h *History) Save(query, identifier string) {
	mut.Lock()
	defer mut.Unlock()

	if _, ok := h.Data[query]; ok {
		if val, ok := h.Data[query][identifier]; ok {
			val.LastUsed = time.Now()
			val.Amount = min(val.Amount+1, 10)
		} else {
			h.Data[query][identifier] = &HistoryData{
				LastUsed: time.Now(),
				Amount:   1,
			}
		}
	} else {
		h.Data[query] = make(map[string]*HistoryData)
		h.Data[query][identifier] = &HistoryData{
			LastUsed: time.Now(),
			Amount:   1,
		}
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)

	err := encoder.Encode(h)
	if err != nil {
		slog.Error("history", "encode", err)
		return
	}

	err = os.MkdirAll(filepath.Dir(common.CacheFile(fmt.Sprintf("%s_history.gob", h.Provider))), 0755)
	if err != nil {
		slog.Error("history", "createdirs", err)
		return
	}

	err = os.WriteFile(common.CacheFile(fmt.Sprintf("%s_history.gob", h.Provider)), b.Bytes(), 0o600)
	if err != nil {
		slog.Error("history", "writefile", err)
	}
}

func (h *History) FindUsage(query, identifier string) (int, time.Time) {
	var longest int
	var usage int
	var lastUsed time.Time

	for k, v := range h.Data {
		if (strings.HasPrefix(query, k) || query == "") && len(k) > longest {
			if n, ok := v[identifier]; ok {
				usage = n.Amount
				longest = len(k)
				lastUsed = n.LastUsed
			}
		}
	}

	return usage, lastUsed
}

func Load(provider string) *History {
	h := History{
		Data:     make(map[string]map[string]*HistoryData),
		Provider: provider,
	}

	f, err := os.ReadFile(common.CacheFile(fmt.Sprintf("%s_history.gob", provider)))
	if err != nil {
		slog.Error("history", "load", err)
	} else {
		decoder := gob.NewDecoder(bytes.NewReader(f))

		err = decoder.Decode(&h)
		if err != nil {
			slog.Error("history", "decoding", err)
		}
	}

	return &h
}
