package configfile

import (
	"errors"
	"fmt"
)

// ErrMalformedConfig means the file is not valid JSON for the schema. The
// error never contains the file contents, which may hold the bot token.
var ErrMalformedConfig = errors.New(malformedMessage)

// ErrInvalidConfig means a value is outside its allowed range.
var ErrInvalidConfig = errors.New("invalid config")

// ErrUnreadableConfig means the file exists but could not be read.
var ErrUnreadableConfig = errors.New(unreadableMessage)

// ErrWriteConfig means the file could not be saved; the previous file is intact.
var ErrWriteConfig = errors.New("config could not be saved")

// ErrUntrustedConfig means the file may not be used to hold a secret
// (symbolic link, foreign owner, writable directory when running as root).
// The error never contains the file contents.
var ErrUntrustedConfig = errors.New(untrustedMessage)

func untrusted(reason string) error {
	return fmt.Errorf(errorDetailFormat, ErrUntrustedConfig, reason)
}

func invalid(detail string) error {
	return fmt.Errorf(errorDetailFormat, ErrInvalidConfig, detail)
}

func writeFailure(detail string, cause error) error {
	// Only the operation and the path error are kept: *os.PathError carries
	// a file path, never file contents.
	return fmt.Errorf("%w: %s: %v", ErrWriteConfig, detail, cause)
}
