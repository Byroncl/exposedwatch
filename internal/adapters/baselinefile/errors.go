package baselinefile

import (
	"errors"
	"fmt"
)

// ErrMalformedBaseline means the file is not valid JSON for the schema or is
// too large. The error never contains the file contents.
var ErrMalformedBaseline = errors.New(malformedMessage)

// ErrUnreadableBaseline means the file exists but could not be read.
var ErrUnreadableBaseline = errors.New(unreadableMessage)

// ErrMissingBaseline means there is no baseline file at the path.
var ErrMissingBaseline = errors.New(missingMessage)

// ErrUntrustedBaseline means the file may have been tampered with: a symbolic
// link, a swapped file, a foreign owner, write access for others, or (as
// root) a directory others can write to. The error never contains the file
// contents.
var ErrUntrustedBaseline = errors.New(untrustedMessage)

// ErrNotABaseline means Save refused to overwrite an existing path that is
// not a regular file or is not a trusted, valid baseline. The error never
// contains the file contents.
var ErrNotABaseline = errors.New(notABaselineMessage)

// ErrWriteBaseline means the file could not be saved; the previous file is intact.
var ErrWriteBaseline = errors.New(writeFailedMessage)

func untrusted(reason string) error {
	return fmt.Errorf(errorDetailFormat, ErrUntrustedBaseline, reason)
}

func writeFailure(detail string, cause error) error {
	// Only the operation and the path error are kept: *os.PathError carries
	// a file path, never file contents.
	return fmt.Errorf(writeFailureFormat, ErrWriteBaseline, detail, cause)
}
