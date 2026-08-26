package data

import (
	"os"
	"path"
	"path/filepath"
)

type FileEntry struct {
	Path      string
	Name      string
	Extension string
	ParentDir string
}

func NewFileEntry(dir string, entry os.DirEntry) FileEntry {
	return FileEntry{
		Path:      filepath.Join(dir, entry.Name()),
		Name:      entry.Name(),
		Extension: path.Ext(entry.Name()),
		ParentDir: dir,
	}
}

func NewFileEntries(dir string, files []os.DirEntry) []FileEntry {
	var fileEntries []FileEntry
	for _, file := range files {
		fileEntries = append(fileEntries, NewFileEntry(dir, file))
	}
	return fileEntries
}
