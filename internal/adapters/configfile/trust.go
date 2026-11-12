package configfile

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Trust is what the loader needs to decide whether a config file may hold a
// secret: who the process runs as, how to inspect paths and who owns a file.
// Everything is injectable for tests.
type Trust struct {
	EffectiveUserID int
	// Lstat inspects a path without following a final symlink.
	Lstat func(path string) (os.FileInfo, error)
	// OwnerOf returns the owning user id, or false when it is unknown.
	OwnerOf func(info os.FileInfo) (int, bool)
}

// SystemTrust inspects the real process and file system.
func SystemTrust() Trust {
	return Trust{EffectiveUserID: os.Geteuid(), Lstat: os.Lstat, OwnerOf: ownerOfFile}
}

func ownerOfFile(info os.FileInfo) (int, bool) {
	details, known := info.Sys().(*syscall.Stat_t)
	if !known {
		return 0, false
	}
	return int(details.Uid), true
}

// checkFile refuses a config file that was swapped, is a symlink or has a
// foreign owner. The caller passes the Lstat result taken before opening and
// the Stat of the already opened descriptor, so a swap between the two shows
// up as a different file.
func (trust Trust) checkFile(linkInfo, openedInfo os.FileInfo) error {
	if linkInfo.Mode()&os.ModeSymlink != 0 || !os.SameFile(linkInfo, openedInfo) {
		return untrusted(symlinkMessage)
	}
	owner, known := trust.OwnerOf(openedInfo)
	if !known || owner != trust.EffectiveUserID {
		return untrusted(ownerMessage)
	}
	return nil
}

// checkParent looks at the directory that holds the file. A directory that
// others can write to lets them replace the file: root refuses it, other
// users get a warning.
func (trust Trust) checkParent(path string) (warning string, err error) {
	directory := filepath.Dir(path)
	info, statErr := trust.Lstat(directory)
	if statErr != nil || info.Mode().Perm()&parentWriteBits == 0 {
		return "", nil
	}
	if trust.EffectiveUserID == rootUserID {
		return "", untrusted(parentWritableMessage)
	}
	return fmt.Sprintf(parentWarningFormat, directory), nil
}
