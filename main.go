package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/digineo/goldflags"
)

func main() {
	fmt.Println(goldflags.Banner("Triax EoC Exporter"))
	log.SetFlags(log.Lshortfile)

	showVersion := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	const l = "%-10s %-50s %s\n"
	fmt.Println("Dependencies\n------------")
	fmt.Printf(l, "main", info.Main.Path, info.Main.Version)
	for _, i := range info.Deps {
		if r := i.Replace; r != nil {
			fmt.Printf(l, "dep", r.Path, r.Version)
			fmt.Printf(l, "  replaces", i.Path, i.Version)
		} else {
			fmt.Printf(l, "dep", i.Path, i.Version)
		}
	}
}
