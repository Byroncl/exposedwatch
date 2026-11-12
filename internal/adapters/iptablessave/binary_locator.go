package iptablessave

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BinaryLocator finds an executable. It checks a fixed list of directories
// first because the iptables tools are in sbin directories that a normal
// user's PATH usually lacks, and falls back to PATH last.
type BinaryLocator struct {
	// Directories are checked in order.
	Directories []string
	// PathLookup is the PATH fallback; nil disables it.
	PathLookup func(name string) (string, error)
}

// NewBinaryLocator returns the production locator.
func NewBinaryLocator() BinaryLocator {
	return BinaryLocator{Directories: DefaultBinaryDirectories, PathLookup: exec.LookPath}
}

// Find returns the full path of the named executable or ErrBinaryNotFound.
// Candidates that the file's owner is not the only one able to change (file or
// directory writable by group or others) are rejected, because the tools may
// run with elevated privileges.
func (locator BinaryLocator) Find(name string) (string, error) {
	for _, directory := range locator.Directories {
		candidate := filepath.Join(directory, name)
		if isSafeExecutable(candidate) {
			return candidate, nil
		}
	}
	if locator.PathLookup != nil {
		if found, err := locator.PathLookup(name); err == nil && isSafeExecutable(found) {
			return found, nil
		}
	}
	return "", fmt.Errorf(binaryMissingFormat, ErrBinaryNotFound, name)
}

// isSafeExecutable reports whether path is a regular executable file that
// neither it nor its directory lets group or others modify.
func isSafeExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	permissions := info.Mode().Perm()
	if permissions&executableModeBits == 0 || permissions&writableByOthersBits != 0 {
		return false
	}
	directoryInfo, err := os.Stat(filepath.Dir(path))
	return err == nil && directoryInfo.Mode().Perm()&writableByOthersBits == 0
}
