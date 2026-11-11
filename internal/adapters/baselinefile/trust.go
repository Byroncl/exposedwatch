package baselinefile

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Trust is what the loader needs to decide whether a baseline file may be
// believed: who the process runs as, how to inspect paths and who owns a
// file. Everything is injectable for tests.
type Trust struct {
	EffectiveUserID int
	// Lstat inspects a path without following a final symlink.
	Lstat func(path string) (os.FileInfo, error)
	// Stat inspects a directory, following symlinks to the real directory.
	Stat func(path string) (os.FileInfo, error)
	// OwnerOf returns the owning user id, or false when it is unknown.
	OwnerOf func(info os.FileInfo) (int, bool)
}

// SystemTrust inspects the real process and file system.
func SystemTrust() Trust {
	return Trust{EffectiveUserID: os.Geteuid(), Lstat: os.Lstat, Stat: os.Stat, OwnerOf: ownerOfFile}
}

func ownerOfFile(info os.FileInfo) (int, bool) {
	details, known := info.Sys().(*syscall.Stat_t)
	if !known {
		return 0, false
	}
	return int(details.Uid), true
}

// checkFile refuses a baseline that was swapped, is a symlink, has a foreign
// owner or can be written by others. The caller passes the Lstat result taken
// before opening and the Stat of the opened descriptor, so a swap between the
// two shows up as a different file.
func (trust Trust) checkFile(linkInfo, openedInfo os.FileInfo) error {
	if linkInfo.Mode()&os.ModeSymlink != 0 || !os.SameFile(linkInfo, openedInfo) {
		return untrusted(symlinkMessage)
	}
	owner, known := trust.OwnerOf(openedInfo)
	if !known || owner != trust.EffectiveUserID {
		return untrusted(ownerMessage)
	}
	if !openedInfo.Mode().IsRegular() {
		return untrusted(notRegularMessage)
	}
	if openedInfo.Mode().Perm()&writeBits != 0 {
		return untrusted(fileWritableMessage)
	}
	return nil
}

// checkParent looks at the directory that holds the file (only that one, not
// its ancestors). A directory that others can write to, or that belongs to
// neither the running user nor root, lets someone else replace the file: root
// refuses it, other users get a warning.
func (trust Trust) checkParent(path string) (warning string, err error) {
	directory := filepath.Dir(path)
	info, statErr := trust.Stat(directory)
	if statErr != nil {
		return "", nil
	}
	if info.Mode().Perm()&writeBits != 0 {
		if trust.EffectiveUserID == rootUserID {
			return "", untrusted(parentWritableMessage)
		}
		return fmt.Sprintf(parentWarningFormat, directory), nil
	}
	owner, known := trust.OwnerOf(info)
	if known && (owner == trust.EffectiveUserID || owner == rootUserID) {
		return "", nil
	}
	if trust.EffectiveUserID == rootUserID {
		return "", untrusted(parentOwnerMessage)
	}
	return fmt.Sprintf(parentOwnerWarningFormat, directory), nil
}
