package cmd

import (
	"flag"
)

type Flags struct {
	ConfigFile string
}

func ParseFlags() Flags {
	config := flag.String("config-file", "~/.axis/config.toml", "path to configuration file")
	flag.Parse()
	return Flags{
		ConfigFile: *config,
	}
}
