package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"git.digineo.de/digineo/triax_eoc_exporter/config"
	"github.com/digineo/goldflags"
)

// DefaultConfigPath points to the default config file location.
// This might be overwritten at build time (using -ldflags).
var DefaultConfigPath = "./config.toml"

func main() {
	fmt.Println(goldflags.Banner("Triax EoC Exporter"))
	log.SetFlags(log.Lshortfile)

	configFile := flag.String("config", DefaultConfigPath, "`PATH` to configuration file")
	showVersion := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *configFile == "" {
		*configFile = DefaultConfigPath
	}

	cfg, err := config.LoadFile(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	const l = "%-10s %-50s %s\n"
	fmt.Println("Dependencies\n------------")
	fmt.Printf(l, "main", info.Main.Path, goldflags.VersionString())
	for _, i := range info.Deps {
		if r := i.Replace; r != nil {
			fmt.Printf(l, "dep", r.Path, r.Version)
			fmt.Printf(l, "  replaces", i.Path, i.Version)
		} else {
			fmt.Printf(l, "dep", i.Path, i.Version)
		}
	}
}
