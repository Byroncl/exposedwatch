// Package baselinefile stores the exposure baseline as a private, trusted JSON file.
package baselinefile

import "os"

// File layout and permissions.
const (
	// MaxFileBytes bounds how much of a baseline file is ever read.
	MaxFileBytes = 4 * 1024 * 1024

	directoryMode os.FileMode = 0o700
	fileMode      os.FileMode = 0o600

	// writeBits are the group and other write bits of a file or directory.
	writeBits os.FileMode = 0o022

	temporaryFilePattern = ".baseline-*.tmp"
	fileIndentation      = "  "
	fileTrailer          = "\n"
)

// Default path resolution.
const (
	rootUserID         = 0
	rootBaselinePath   = "/var/lib/exposedwatch/baseline.json"
	envXDGStateHome    = "XDG_STATE_HOME"
	homeStateDirectory = ".local/state"
	applicationFolder  = "exposedwatch"
	baselineFileName   = "baseline.json"
	noHomeMessage      = "cannot locate a state directory: no home directory and no XDG_STATE_HOME"
)

// Messages. None of them include file contents.
const (
	malformedMessage         = "baseline file is not valid JSON for this schema (contents are not shown)"
	oversizedMessage         = "baseline file is larger than the allowed size"
	unreadableMessage        = "baseline file could not be read"
	missingMessage           = "baseline file does not exist"
	untrustedMessage         = "baseline file is not trusted"
	writeFailedMessage       = "baseline could not be saved"
	createdAtInvalidText     = "created_at must be an RFC 3339 timestamp"
	errorDetailFormat        = "%w: %s"
	symlinkMessage           = "it is a symbolic link or changed while it was opened"
	ownerMessage             = "it is not owned by the user running exposedwatch"
	fileWritableMessage      = "it can be written by group or others"
	parentWritableMessage    = "its directory can be written by group or others"
	parentOwnerMessage       = "its directory is owned by another user"
	notRegularMessage        = "it is not a regular file"
	notABaselineMessage      = "existing file is not a baseline"
	existingUnusableText     = "it could not be read as a trusted baseline"
	parentOwnerWarningFormat = "directory %s is owned by another user, who could replace the baseline file"
	parentWarningFormat      = "directory %s can be written by group or others, which would let them replace the baseline file"
	encodeMessage            = "baseline could not be encoded"
	createDirectoryMessage   = "baseline directory could not be created"
	writeTemporaryMessage    = "temporary baseline file could not be written"
	replaceFileMessage       = "baseline file could not be replaced"
	writeFailureFormat       = "%w: %s: %v"
)
