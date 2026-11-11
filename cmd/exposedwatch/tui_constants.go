package main

import (
	"os"
	"time"
)

// Limits of the tui settings wiring.
const (
	// alertTestTimeout bounds the whole test message, so a stuck network
	// cannot keep the tui form busy.
	alertTestTimeout = 15 * time.Second
)

// Directory permission checks.
const (
	// directoryWriteMask holds the group and other write bits of a directory.
	directoryWriteMask os.FileMode = 0o022
	rootUserID                     = 0
)

// Messages of the tui settings wiring. None of them may contain a token.
const (
	testMessageNeedsCredentials = "a test message needs a bot token and a chat ID"
	directoryWritableMessage    = "the config directory is writable by group or others; fix its permissions first"
	saveRefusedFormat           = "the config file was not changed: %w"
)
