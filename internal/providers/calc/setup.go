package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/pkg/pb/pb"
)

var (
	Name       = "calc"
	NamePretty = "Calculator/Unit-Conversion"
	config     *Config
)

type Config struct {
	common.Config `koanf:",squash"`
	MaxItems      int `koanf:"max_items" desc:"max amount of calculation history items" default:"100"`
}

func init() {
	config = &Config{
		Config:   common.Config{},
		MaxItems: 100,
	}

	common.LoadConfig(Name, config)

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

func Cleanup(qid uint32) {}

func Activate(qid uint32, identifier, action string, arguments string) {
}

func Query(qid uint32, iid uint32, query string) []*pb.QueryResponse_Item {
	start := time.Now()
	entries := []*pb.QueryResponse_Item{}

	cmd := exec.Command("qalc", "-t", query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error(Name, "query", err)
	} else {
		e := &pb.QueryResponse_Item{
			Identifier: "",
			Text:       string(out),
			Subtext:    query,
			Provider:   Name,
			Type:       pb.QueryResponse_REGULAR,
		}

		entries = append(entries, e)
	}

	slog.Info(Name, "queryresult", len(entries), "time", time.Since(start))

	return entries
}
