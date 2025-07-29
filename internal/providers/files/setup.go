package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/util"
)

var (
	paths        []string
	results      map[uint32][]string
	resultsMutex sync.Mutex
)

var (
	Name       = "files"
	NamePretty = "Files"
)

const (
	ActionOpen    = "open"
	ActionOpenDir = "opendir"
)

func init() {
	loadConfig()
}

func PrintDoc() {
	fmt.Printf("### %s\n", Name)
	fmt.Println("Search files and folders.")
	fmt.Println()
	util.PrintConfig(Config{})
}

func Cleanup(qid uint32) {
	slog.Info(Name, "cleanup", qid)
	resultsMutex.Lock()
	delete(results, qid)
	resultsMutex.Unlock()
}

func Activate(qid uint32, identifier, action string) {
	i, err := strconv.Atoi(identifier)
	if err != nil {
		slog.Error(Name, "activate", err)
		return
	}

	switch action {
	// TODO: find out if it needs to be opened in a terminal, see Walker
	case ActionOpen:
		cmd := exec.Command("sh", "-c", common.WrapWithPrefix(config.LaunchPrefix, fmt.Sprintf("xdg-open '%s'", paths[i])))
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		err := cmd.Start()
		if err != nil {
			slog.Error(Name, "actionopen", err)
		}

		go func() {
			cmd.Wait()
		}()
	default:
		slog.Error(Name, "nosuchaction", action)
	}
}

func Query(qid uint32, text string) []common.Entry {
	start := time.Now()
	entries := []common.Entry{}

	var toMatch []string

	v, ok := results[qid]
	if ok {
		toMatch = v
		slog.Info(Name, "query", "resuming", "files", len(toMatch))
	} else {
		toMatch = paths
	}

	resultsMutex.Lock()
	results[qid] = []string{}
	resultsMutex.Unlock()

	for k, v := range toMatch {
		common.FuzzyScore(text, v)

		i := strconv.Itoa(k)

		e := common.Entry{
			Identifier: i,
			Text:       v,
			SubText:    "",
			Provider:   Name,
		}

		var match string
		var ok bool

		if text != "" {
			e.Fuzzy = &common.FuzzyMatchInfo{
				Field: "text",
			}

			match, e.Score, e.Fuzzy.Pos, e.Fuzzy.Start, ok = calcScore(text, v)

			if ok && match != e.Text {
				e.SubText = match
				e.Fuzzy.Field = "text"
			}
		}

		if e.Score > 0 || text == "" {
			resultsMutex.Lock()
			results[qid] = append(results[qid], v)
			resultsMutex.Unlock()

			entries = append(entries, e)
		}
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))
	return entries
}

func Load() {
	results = make(map[uint32][]string)
	start := time.Now()
	paths = []string{}
	home, _ := os.UserHomeDir()
	cmd := exec.Command("fd", ".", home, "--ignore-vcs", "--type", "file", "--type", "directory")

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error(Name, "files", err)
		os.Exit(1)
	}

	for v := range bytes.Lines(out) {
		if len(v) > 0 {
			paths = append(paths, strings.TrimSpace(string(v)))
		}
	}

	slog.Info(Name, "files", len(paths), "time", time.Since(start))
}

func calcScore(q string, d string) (string, int, *[]int, int, bool) {
	var scoreRes int
	var posRes *[]int
	var startRes int
	var match string
	var modifier int

	score, pos, start := common.FuzzyScore(q, d)

	if score > scoreRes {
		scoreRes = score
		posRes = pos
		startRes = start
		match = d
		modifier = 0
	}

	if scoreRes == 0 {
		return "", 0, nil, 0, false
	}

	scoreRes = max(scoreRes-min(modifier*10, 50)-startRes, 10)

	return match, scoreRes, posRes, startRes, true
}

func EntryToString(e common.Entry) string {
	var start int
	var field string

	positions := []string{}

	if e.Fuzzy != nil {
		if e.Fuzzy.Pos != nil {
			for _, num := range *e.Fuzzy.Pos {
				positions = append(positions, strconv.Itoa(num))
			}
		}

		start = e.Fuzzy.Start
		field = e.Fuzzy.Field
	}

	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;%d;%s", e.Provider, e.Identifier, e.Text, "", "", strings.Join(positions, ","), start, field)
}
