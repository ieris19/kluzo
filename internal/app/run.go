package app

import (
	"git.ierislabs.dev/ieris19/kluzo/internal/config"
	"git.ierislabs.dev/ieris19/kluzo/internal/container"
	"git.ierislabs.dev/ieris19/kluzo/internal/data"
	"git.ierislabs.dev/ieris19/kluzo/internal/parser"
)

// Run executes the scan -> parse -> check pipeline and returns the assembled
// report. The report includes all the relevant information about this run
func Run(cfg config.Config) data.UpdateReport {
	parser.SetAliases(cfg.Registry.Aliases)
	containerFiles, scanErrors := parser.GetAllFiles(cfg.Scanner, parser.SupportedExtensions)

	report := data.UpdateReport{
		Outdated: []data.Update{},
		Updated:  []data.Update{},
		Frozen:   []data.ContainerDefinition{},
		Skipped:  []data.ContainerError{},
		Errors:   scanErrors,
	}

	containerDefinitions := parser.ParseContainerFiles(containerFiles, &report)

	for _, cnt := range containerDefinitions {
		// Frozen containers are skipped entirely
		if cnt.Pin == data.PinFreeze {
			report.Frozen = append(report.Frozen, cnt)
			continue
		}

		update, err := container.CheckUpdate(cnt)
		// Problem
		if err != nil {
			report.Errors = append(report.Errors, data.ContainerError{
				File:  cnt.File,
				Name:  cnt.Name,
				Stage: data.CheckStage,
				Err:   err,
			})
			continue
		}

		// Update available
		if update.Upgradeable {
			report.Outdated = append(report.Outdated, update)
			continue
		}

		// No updates available
		report.Updated = append(report.Updated, update)
	}

	return report
}
