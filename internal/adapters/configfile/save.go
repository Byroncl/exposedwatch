package configfile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// renameFunc is os.Rename, replaceable in tests to simulate failures.
type renameFunc func(oldPath, newPath string) error

// Save validates the config and writes it atomically with mode 0600 inside
// a 0700 directory. A failed save leaves the previous file untouched, and
// saving also repairs loose permissions of an older file.
func Save(path string, config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	content, err := json.MarshalIndent(config, "", fileIndentation)
	if err != nil {
		return writeFailure(encodeMessage, err)
	}
	content = append(content, fileTrailer...)
	if err := os.MkdirAll(filepath.Dir(path), directoryMode); err != nil {
		return writeFailure(createDirectoryMessage, err)
	}
	return writeAtomically(path, content, os.Rename)
}

// writeAtomically writes to a private temporary file in the same directory,
// flushes it to disk and renames it over the destination.
func writeAtomically(path string, content []byte, rename renameFunc) (writeErr error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), temporaryFilePattern)
	if err != nil {
		return writeFailure(writeTemporaryMessage, err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if writeErr != nil {
			_ = temporary.Close()
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(fileMode); err != nil {
		return writeFailure(writeTemporaryMessage, err)
	}
	if _, err := temporary.Write(content); err != nil {
		return writeFailure(writeTemporaryMessage, err)
	}
	if err := temporary.Sync(); err != nil {
		return writeFailure(writeTemporaryMessage, err)
	}
	if err := temporary.Close(); err != nil {
		return writeFailure(writeTemporaryMessage, err)
	}
	if err := rename(temporaryPath, path); err != nil {
		return writeFailure(replaceFileMessage, err)
	}
	return nil
}
