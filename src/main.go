package main

import (
    "flag"
    "fmt"
    "os"

    "git.ierislabs.dev/update-link/config"
    "git.ierislabs.dev/update-link/container"
    "git.ierislabs.dev/update-link/files"
    "git.ierislabs.dev/update-link/parser"
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
    containerFiles := files.GetAllFiles(cfg.Scanner.Directories, parser.SupportedExtensions)
    containerDefinitions := parser.ParseContainerFiles(containerFiles)
    // Print parsed container definitions
    var availableUpdates []container.Update
    var upToDate []container.Update

    for _, cnt := range containerDefinitions {
        update, err := container.CheckUpdate(cnt)
        if err != nil {
            fmt.Printf("Error checking updates for %s: %v\n", cnt.File.Path, err)
            continue
        }

        if update.Upgradeable {
            availableUpdates = append(availableUpdates, update)
        } else {
            upToDate = append(upToDate, update)
        }
    }
    report(upToDate, availableUpdates)
}

func report(upToDate []container.Update, availableUpdates []container.Update) {
    if len(availableUpdates) > 0 {
        fmt.Println("Available Updates:")
        for _, update := range availableUpdates {
            fmt.Printf("- %s: Current version %s, Latest version %s\n", update.ContainerDefinition.Name, update.ContainerDefinition.Version.String(), update.LatestVersion.String())
        }
    }
    if len(upToDate) > 0 {
        fmt.Println("\nUp-to-date Containers:")
        for _, update := range upToDate {
            fmt.Printf("- %s: Version %s\n", update.ContainerDefinition.Name, update.ContainerDefinition.Version.String())
        }
    }
}
