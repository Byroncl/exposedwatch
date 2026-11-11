// Package configfile stores the user configuration as a private JSON file.
package configfile

import "os"

// File layout and permissions.
const (
	// CurrentVersion is the only schema version this build understands.
	CurrentVersion = 1
	// MaxFileBytes bounds how much of a config file is ever read.
	MaxFileBytes = 64 * 1024

	directoryMode os.FileMode = 0o700
	fileMode      os.FileMode = 0o600
	// groupAndOtherBits are the permission bits that make a secret file leaky.
	groupAndOtherBits os.FileMode = 0o077

	// parentWriteBits are the group and other write bits of a directory.
	parentWriteBits os.FileMode = 0o022

	temporaryFilePattern = ".config-*.tmp"
	fileIndentation      = "  "
	fileTrailer          = "\n"
)

// Environment variables that override the file. They let the bot token live
// outside the disk (for example in a systemd EnvironmentFile).
const (
	EnvTelegramToken  = "EXPOSEDWATCH_TELEGRAM_TOKEN"
	EnvTelegramChatID = "EXPOSEDWATCH_TELEGRAM_CHAT_ID"
)

// Default path resolution.
const (
	rootUserID          = 0
	rootConfigPath      = "/etc/exposedwatch/config.json"
	envXDGConfigHome    = "XDG_CONFIG_HOME"
	homeConfigDirectory = ".config"
	applicationFolder   = "exposedwatch"
	configFileName      = "config.json"
)

// Default and bounded settings.
const (
	defaultMinSeverity     = "high"
	defaultCooldownMinutes = 60
	defaultIntervalSeconds = 60

	MinCooldownMinutes = 0
	MaxCooldownMinutes = 7 * 24 * 60
	MinIntervalSeconds = 10
	MaxIntervalSeconds = 24 * 60 * 60
)

// Token masking.
const (
	tokenSeparator = ":"
	maskedSuffix   = "****"
)

// Messages.
const (
	malformedMessage       = "config file is not valid JSON for this schema (contents are not shown)"
	oversizedMessage       = "config file is larger than the allowed size"
	unreadableMessage      = "config file could not be read"
	invalidVersionFormat   = "unsupported config version %d (this build understands version %d)"
	invalidSeverityMessage = "alerts.min_severity must be one of info, low, medium, high, critical"
	invalidCooldownFormat  = "alerts.cooldown_minutes must be between %d and %d"
	invalidIntervalFormat  = "alerts.interval_seconds must be between %d and %d"
	looseModeFormat        = "config file %s is accessible by other users (mode %04o); saving it again restricts it to its owner"
	noHomeMessage          = "cannot locate a config directory: no home directory and no XDG_CONFIG_HOME"
	errorDetailFormat      = "%w: %s"
	untrustedMessage       = "config file is not trusted"
	symlinkMessage         = "it is a symbolic link or changed while it was opened"
	ownerMessage           = "it is not owned by the user running exposedwatch"
	parentWritableMessage  = "its directory can be written by group or others"
	parentWarningFormat    = "directory %s can be written by group or others, which would let them replace the config file"
	encodeMessage          = "config could not be encoded"
	createDirectoryMessage = "config directory could not be created"
	writeTemporaryMessage  = "temporary config file could not be written"
	replaceFileMessage     = "config file could not be replaced"
)
