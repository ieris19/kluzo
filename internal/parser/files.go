package parser

import (
	"io/fs"
	"os"
	"path/filepath"

	"git.ierislabs.dev/ieris19/kluzo/internal/config"
	"git.ierislabs.dev/ieris19/kluzo/internal/data"
)

func GetAllFiles(settings config.ScannerConfig, allowedExtensions []string) ([]data.FileEntry, []data.ContainerError) {
	var containerFiles []data.FileEntry
	var errs []data.ContainerError

	for _, dir := range settings.Directories {
		walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			// Returning nil here means continue to the next file
			// Returning an error WalkDir doesn't expect aborts the search
			if err != nil {
				errs = append(errs, data.ContainerError{File: data.FileEntry{Path: path}, Stage: data.ScanStage, Err: err})
				return nil
			}
			// Patterns match against the path relative to the scan root
			rel, relErr := filepath.Rel(dir, path)
			if relErr != nil {
				rel = path
			}
			excluded := isExcluded(rel, settings.Exclude)

			if d.IsDir() {
				if path != dir && excluded {
					// WalkDir will swallow this error and won't propagate, safe to use here
					return fs.SkipDir
				}
				return nil
			}
			if excluded || !isSupportedExtension(filepath.Ext(d.Name()), allowedExtensions) {
				return nil
			}
			containerFiles = append(containerFiles, data.NewFileEntry(filepath.Dir(path), d))
			return nil
		})
		// Should not trigger with current code, kept in case a future change aborts the search
		if walkErr != nil {
			errs = append(errs, data.ContainerError{File: data.FileEntry{Path: dir}, Stage: data.ScanStage, Err: walkErr})
		}
	}

	return containerFiles, errs
}

func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
