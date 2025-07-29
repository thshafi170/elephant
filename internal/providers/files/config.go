package main

import "github.com/abenz1267/elephant/internal/common"

type Config struct {
	common.Config `koanf:",squash"`
	LaunchPrefix  string `koanf:"launch_prefix" desc:"overrides the default app2unit or uwsm prefix, if set. 'CLEAR' to not prefix." default:""`
}

var config *Config

func loadConfig() {
	config = &Config{
		Config:       common.Config{},
		LaunchPrefix: "",
	}

	common.LoadConfig(Name, config)
}
