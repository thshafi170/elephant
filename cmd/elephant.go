package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/abenz1267/elephant/internal/comm"
	"github.com/abenz1267/elephant/internal/common"
	"github.com/abenz1267/elephant/internal/providers"
	"github.com/abenz1267/elephant/internal/util"
	"github.com/urfave/cli/v3"
)

func main() {
	var config string

	cmd := &cli.Command{
		Name:                   "Elephant",
		Usage:                  "Data provider and executor",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:    "generatedoc",
				Aliases: []string{"d"},
				Usage:   "generates a markdown documentation",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					providers.Load()
					util.GenerateDoc()
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "",
				Destination: &config,
				Usage:       "config folder location",
				Action: func(ctx context.Context, cmd *cli.Command, val string) error {
					common.SetExplicitDir(val)
					return nil
				},
			},
		},
		Action: func(context.Context, *cli.Command) error {
			start := time.Now()
			providers.Load()
			providers.Setup()
			slog.Info("elephant", "startup", time.Since(start))
			comm.StartListen()

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
