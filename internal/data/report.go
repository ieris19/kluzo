package data

import "fmt"

type Update struct {
	Definition    ContainerDefinition
	LatestVersion SemanticVersion
	Upgradeable   bool
	// Pinned is true when no update is available within the version pin,
	// but a newer version exists outside of it (e.g. a new major release).
	Pinned bool
}

type Stage string

const (
	ParseStage Stage = "parse"
	CheckStage Stage = "check"
	ScanStage  Stage = "scan"
)

type ContainerError struct {
	File  FileEntry
	Name  string
	Stage Stage
	Err   error
}

func (e ContainerError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s (%s): %v", e.Name, e.File.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.File.Path, e.Err)
}

func (e ContainerError) Unwrap() error {
	return e.Err
}

type UpdateReport struct {
	Outdated []Update
	Updated  []Update
	Frozen   []ContainerDefinition
	Skipped  []ContainerError
	Errors   []ContainerError
}
