package baselinefile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/Byroncl/exposedwatch/internal/ports"
)

// Load reads and validates the baseline. A missing file is ErrMissingBaseline.
// A symlink, a file swapped while it is opened, a foreign owner, a file
// others can write to, or (as root) a directory others can write to is
// refused as ErrUntrustedBaseline: the baseline decides what counts as
// accepted, so a tampered file would hide new ports. Other users only get a
// warning about the directory.
func (store Store) Load(path ports.BaselinePath) (ports.BaselineLoad, error) {
	content, warnings, err := store.readTrusted(string(path))
	if err != nil {
		return ports.BaselineLoad{}, err
	}
	baseline, err := decodeDocument(content)
	if err != nil {
		return ports.BaselineLoad{}, err
	}
	return ports.BaselineLoad{Baseline: baseline, Warnings: warnings}, nil
}

// readTrusted checks the file with Lstat, opens it, checks that the opened
// descriptor is the same non-symlink file with an acceptable owner, and reads
// through that descriptor, so the file cannot be swapped after the checks.
func (store Store) readTrusted(path string) ([]byte, []string, error) {
	linkInfo, err := store.trust.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf(errorDetailFormat, ErrMissingBaseline, err)
		}
		return nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableBaseline, err)
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 {
		return nil, nil, untrusted(symlinkMessage)
	}
	if !linkInfo.Mode().IsRegular() {
		return nil, nil, untrusted(notRegularMessage)
	}
	// The directory is checked before the file is opened, and the open never
	// blocks (a FIFO would hang a blocking open) and never follows a symlink.
	var warnings []string
	warning, err := store.trust.checkParent(path)
	if err != nil {
		return nil, nil, err
	}
	if warning != "" {
		warnings = append(warnings, warning)
	}
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return nil, nil, untrusted(symlinkMessage)
		}
		return nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableBaseline, err)
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableBaseline, err)
	}
	if err := store.trust.checkFile(linkInfo, openedInfo); err != nil {
		return nil, nil, err
	}
	content, err := io.ReadAll(io.LimitReader(file, MaxFileBytes+1))
	if err != nil {
		return nil, nil, fmt.Errorf(errorDetailFormat, ErrUnreadableBaseline, err)
	}
	if len(content) > MaxFileBytes {
		return nil, nil, fmt.Errorf(errorDetailFormat, ErrMalformedBaseline, oversizedMessage)
	}
	return content, warnings, nil
}
