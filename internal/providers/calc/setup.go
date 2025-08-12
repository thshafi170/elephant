package main

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "calc"
	NamePretty = "Calculator/Unit-Conversion"
	config     *Config
)

const (
	ActionCopy   = "copy"
	ActionSave   = "save"
	ActionDelete = "delete"
)

type Config struct {
	common.Config `koanf:",squash"`
	MaxItems      int `koanf:"max_items" desc:"max amount of calculation history items" default:"100"`
}

type HistoryItem struct {
	Identifier string
	Input      string
	Result     string
}

var history = []HistoryItem{}

var (
	resultMutex sync.Mutex
	results     = make(map[uint32]map[string]*pb.QueryResponse_Item)
)

func init() {
	config = &Config{
		Config:   common.Config{},
		MaxItems: 100,
	}

	common.LoadConfig(Name, config)

	loadHist()

	// this is to update exchange rate data
	cmd := exec.Command("qalc", "-e", "1+1")
	err := cmd.Start()
	if err != nil {
		slog.Error(Name, "init", err)
	} else {
		go func() {
			cmd.Wait()
		}()
	}
}

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Calculator/Unit-Conversion with history.")
	fmt.Println()
}

func Cleanup(qid uint32) {
	resultMutex.Lock()
	delete(results, qid)
	resultMutex.Unlock()
}

func Activate(qid uint32, identifier, action string, arguments string) {
	var item *pb.QueryResponse_Item
	var result string
	var createHistoryItem bool

	for _, v := range results {
		if i, ok := v[identifier]; ok {
			item = i
			result = i.Text
			createHistoryItem = true
		}
	}

	for _, v := range history {
		if v.Identifier == identifier {
			result = v.Result
			createHistoryItem = false
		}
	}

	if result == "" {
		slog.Error(Name, "activation", "item not found")
		return
	}

	switch action {
	case ActionCopy:
		cmd := exec.Command("wl-copy", result)

		err := cmd.Start()
		if err != nil {
			slog.Error(Name, "actioncopy", err)
		} else {
			go func() {
				cmd.Wait()
			}()
		}

		if createHistoryItem {
			saveToHistory(item)
		}

		Cleanup(qid)
	case ActionSave:
		if createHistoryItem {
			saveToHistory(item)
		}

		Cleanup(qid)
	case ActionDelete:
		i := 0

		for k, v := range history {
			if v.Identifier == identifier {
				i = k
				break
			}
		}

		history = append(history[:i], history[i+1:]...)

		saveHist()
	}
}

func saveToHistory(item *pb.QueryResponse_Item) {
	h := HistoryItem{
		Identifier: item.Identifier,
		Input:      item.Subtext,
		Result:     item.Text,
	}

	history = append([]HistoryItem{h}, history...)

	saveHist()
}

func Query(qid uint32, iid uint32, query string) []*pb.QueryResponse_Item {
	start := time.Now()

	if _, ok := results[qid]; !ok {
		results[qid] = make(map[string]*pb.QueryResponse_Item)
	}

	entries := []*pb.QueryResponse_Item{}

	if query != "" {
		cmd := exec.Command("qalc", "-t", query)
		out, err := cmd.CombinedOutput()
		if err != nil {
			slog.Error(Name, "query", err)
		} else {
			md5 := md5.Sum([]byte(query))
			md5str := hex.EncodeToString(md5[:])

			e := &pb.QueryResponse_Item{
				Identifier: md5str,
				Text:       strings.TrimSpace(string(out)),
				Subtext:    query,
				Provider:   Name,
				Score:      int32(config.MaxItems) + 1,
				Type:       pb.QueryResponse_REGULAR,
			}

			resultMutex.Lock()
			results[qid][md5str] = e
			resultMutex.Unlock()

			entries = append(entries, e)
		}
	}

	for k, v := range history {
		e := &pb.QueryResponse_Item{
			Identifier: v.Identifier,
			Text:       v.Result,
			Score:      int32(config.MaxItems - k),
			Subtext:    v.Input,
			Provider:   Name,
			Type:       pb.QueryResponse_REGULAR,
		}

		entries = append(entries, e)
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))

	return entries
}

func loadHist() {
	file := common.CacheFile(fmt.Sprintf("%s.gob", Name))

	if common.FileExists(file) {
		f, err := os.ReadFile(file)
		if err != nil {
			slog.Error(Name, "history", err)
		} else {
			decoder := gob.NewDecoder(bytes.NewReader(f))

			err = decoder.Decode(&history)
			if err != nil {
				slog.Error(Name, "decoding", err)
			}
		}
	}
}

func saveHist() {
	if len(history) > config.MaxItems {
		history = history[:config.MaxItems]
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)

	err := encoder.Encode(history)
	if err != nil {
		slog.Error("history", "encode", err)
		return
	}

	err = os.MkdirAll(filepath.Dir(common.CacheFile(fmt.Sprintf("%s.gob", Name))), 0755)
	if err != nil {
		slog.Error("history", "createdirs", err)
		return
	}

	err = os.WriteFile(common.CacheFile(fmt.Sprintf("%s.gob", Name)), b.Bytes(), 0o600)
	if err != nil {
		slog.Error("history", "writefile", err)
	}
}
