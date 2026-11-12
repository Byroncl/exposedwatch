package domain

// Baseline format limits.
const (
	// BaselineVersion is the only baseline schema version this build understands.
	BaselineVersion = 1
	// MaxBaselineEntries bounds how many ports a baseline may hold.
	MaxBaselineEntries = 20000
	// MaxToolVersionLength bounds the free-text tool version stored in a baseline.
	MaxToolVersionLength = 64

	// MinPortNumber and MaxPortNumber delimit valid TCP and UDP port numbers.
	MinPortNumber = 1
	MaxPortNumber = 65535

	// WildcardBindAddress stands in for a missing bind address, which means all interfaces.
	WildcardBindAddress = "0.0.0.0"
)

// Baseline error messages. None of them include the offending value, so a
// corrupt or hostile file is never echoed back.
const (
	invalidBaselineMessage       = "invalid baseline"
	baselineVersionFormat        = "unsupported baseline version %d (this build understands version %d)"
	baselineCreatedAtMessage     = "created_at is missing"
	baselineToolVersionMessage   = "tool_version is too long or contains control characters"
	baselineTooManyEntriesFormat = "baseline holds more than %d entries"
	baselineEntryFormat          = "entry %d: %w"
	entryProtocolMessage         = "protocol must be tcp or udp"
	entryPortFormat              = "port must be between %d and %d"
	entryBindMessage             = "bind_address must be an IP address"
	entrySourceMessage           = "source must be host or docker"
)
