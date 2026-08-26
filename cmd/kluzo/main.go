package main

import (
	"flag"
	"fmt"
	"os"

	"git.ierislabs.dev/ieris19/kluzo/internal/app"
	"git.ierislabs.dev/ieris19/kluzo/internal/config"
	"git.ierislabs.dev/ieris19/kluzo/internal/output"
)

func main() {
	configPath := flag.String("config", "", "path to config.toml (default path if empty)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	report := app.Run(cfg)
	output.ReportStdout(report)
}
