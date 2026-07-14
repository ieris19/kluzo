package output

import (
    "fmt"

    "git.ierislabs.dev/update-link/data"
)

func containerId(container data.ContainerDefinition) string {
    return fmt.Sprintf("%s#%s", container.File.Path, container.Name)
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
            fmt.Printf("- %s: %s\n", containerId(update.Definition), update.Definition.Version.String())
        }
    }
    if len(containers.Errors) > 0 {
        fmt.Println("\nUp-to-date Containers:")
        for _, err := range containers.Errors {
            fmt.Printf("- %s: %s\n", containerId(err.Definition), err.Message)
        }
    }
    fmt.Printf("\nOutdated: %d | Up to Date: %d | Total: %d\n", len(containers.Outdated), len(containers.Updated), len(containers.Errors)+len(containers.Updated)+len(containers.Outdated))
}
