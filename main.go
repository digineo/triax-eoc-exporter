package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/digineo/triax-eoc-exporter/exporter"
	"github.com/digineo/triax-eoc-exporter/triax"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// DefaultConfigPath points to the default config file location.
// This might be overwritten at build time (using -ldflags).
var DefaultConfigPath = "./config.toml"

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	log.SetFlags(log.Lshortfile)

	listenAddress := kingpin.Flag(
		"web.listen-address",
		"Address on which to expose metrics and web interface.",
	).Default(":9809").String()

	configFile := kingpin.Flag(
		"web.config",
		"Path to config.toml that contains all the targets.",
	).Default(DefaultConfigPath).String()

	verbose := kingpin.Flag(
		"verbose",
		"Increase verbosity",
	).Bool()

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	cfg, err := exporter.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	triax.Verbose = *verbose
	cfg.Start(*listenAddress, version)
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	const l = "%-10s %-50s %s\n"
	fmt.Println("Dependencies\n------------")
	fmt.Printf(l, "main", info.Main.Path, version)
	for _, i := range info.Deps {
		if r := i.Replace; r != nil {
			fmt.Printf(l, "dep", r.Path, r.Version)
			fmt.Printf(l, "  replaces", i.Path, i.Version)
		} else {
			fmt.Printf(l, "dep", i.Path, i.Version)
		}
	}
}
