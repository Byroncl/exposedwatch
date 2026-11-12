package baselinefile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// renameFunc is os.Rename, replaceable in tests to simulate failures.
type renameFunc func(oldPath, newPath string) error

// Save validates the baseline and writes it atomically with mode 0600 inside
// a 0700 directory. A failed save leaves the previous file untouched, and
// saving also repairs loose permissions of an older file.
func (store Store) Save(path ports.BaselinePath, baseline domain.Baseline) error {
	content, err := encodeDocument(baseline)
	if err != nil {
		return err
	}
	if err := store.checkExisting(string(path)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(string(path)), directoryMode); err != nil {
		return writeFailure(createDirectoryMessage, err)
	}
	return writeAtomically(string(path), content, os.Rename)
}

// writeAtomically writes to a private temporary file (created exclusively
// with mode 0600) in the same directory, flushes it to disk and renames it
// over the destination. On any failure the temporary file is removed.
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

// checkExisting refuses to replace anything that is not a baseline: the
// target must be missing, or a regular file (never a symlink) that loads as a
// trusted, valid baseline. A forced store only skips that last content check.
func (store Store) checkExisting(path string) error {
	info, err := store.trust.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf(errorDetailFormat, ErrNotABaseline, existingUnusableText)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf(errorDetailFormat, ErrNotABaseline, notRegularMessage)
	}
	if store.force {
		return nil
	}
	content, _, err := store.readTrusted(path)
	if err == nil {
		_, err = decodeDocument(content)
	}
	if err != nil {
		return fmt.Errorf(errorDetailFormat, ErrNotABaseline, existingUnusableText)
	}
	return nil
}
