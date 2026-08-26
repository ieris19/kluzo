package main

import (
	"flag"
	"fmt"
	"os"

	"git.ierislabs.dev/ieris19/kluzo/internal/config"
	"git.ierislabs.dev/ieris19/kluzo/internal/container"
	"git.ierislabs.dev/ieris19/kluzo/internal/data"
	"git.ierislabs.dev/ieris19/kluzo/internal/output"
	"git.ierislabs.dev/ieris19/kluzo/internal/parser"
)

func main() {
	configPath := flag.String("config", "", "path to config.toml (default path if empty)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	parser.SetAliases(cfg.Registry.Aliases)
	containerFiles, err := parser.GetAllFiles(cfg.Scanner, parser.SupportedExtensions)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

	}
	containerDefinitions, parseErrors := parser.ParseContainerFiles(containerFiles)
	// Print parsed container definitions
	cntReport := data.UpdateReport{
		Outdated: []data.Update{},
		Updated:  []data.Update{},
		Frozen:   []data.ContainerDefinition{},
		Errors:   parseErrors,
	}

	for _, cnt := range containerDefinitions {
		// Frozen containers are skipped entirely
		if cnt.Pin == data.PinFreeze {
			cntReport.Frozen = append(cntReport.Frozen, cnt)
			continue
		}

		update, err := container.CheckUpdate(cnt)
		// Problem
		if err != nil {
			cntReport.Errors = append(cntReport.Errors, data.ContainerError{
				File:  cnt.File,
				Name:  cnt.Name,
				Stage: data.CheckStage,
				Err:   err,
			})
			continue
		}

		// Update available
		if update.Upgradeable {
			cntReport.Outdated = append(cntReport.Outdated, update)
			continue
		}

		// No updates available
		cntReport.Updated = append(cntReport.Updated, update)
	}
	output.ReportStdout(cntReport)
}
