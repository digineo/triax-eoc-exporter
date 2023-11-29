package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/digineo/triax-eoc-exporter/exporter"
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

	initLogger(*verbose)

	cfg, err := exporter.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg.Start(*listenAddress, version)
}

func initLogger(verbose bool) {
	opts := slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time from the output
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}

	if verbose {
		opts.Level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)

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
