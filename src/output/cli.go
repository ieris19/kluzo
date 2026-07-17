package output

import (
    "fmt"

    "git.ierislabs.dev/update-link/data"
)

func containerId(container data.ContainerDefinition) string {
    return fmt.Sprintf("%s#%s", container.File.Path, container.Name)
}

func errorId(err data.ContainerError) string {
    if err.Name != "" {
        return fmt.Sprintf("%s#%s", err.File.Path, err.Name)
    }
    return err.File.Path
}

func ReportStdout(containers data.UpdateReport) {
    if len(containers.Outdated) > 0 {
        fmt.Println("Available Updates:")
        for _, update := range containers.Outdated {
            fmt.Printf("- %s: Current version %s, Latest version %s\n", containerId(update.Definition), update.Definition.Version.String(), update.LatestVersion.String())
        }
    }
    if len(containers.Updated) > 0 {
        fmt.Println("\nUp-to-date Containers:")
        for _, update := range containers.Updated {
            pinned := ""
            if update.Pinned {
                pinned = " (pinned)"
            }
            fmt.Printf("- %s: %s%s\n", containerId(update.Definition), update.Definition.Version.String(), pinned)
        }
    }
    if len(containers.Frozen) > 0 {
        fmt.Println("\nFrozen Containers:")
        for _, definition := range containers.Frozen {
            fmt.Printf("- %s: %s\n", containerId(definition), definition.Version.String())
        }
    }
    if len(containers.Errors) > 0 {
        fmt.Println("\nErrors:")
        for _, err := range containers.Errors {
            fmt.Printf("- %s [%s]: %v\n", errorId(err), err.Stage, err.Err)
        }
    }
    total := len(containers.Errors) + len(containers.Updated) + len(containers.Outdated) + len(containers.Frozen)
    fmt.Printf("\nOutdated: %d | Up to Date: %d | Frozen: %d | Errors: %d | Total: %d\n", len(containers.Outdated), len(containers.Updated), len(containers.Frozen), len(containers.Errors), total)
}
