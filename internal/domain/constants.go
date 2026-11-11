package domain

// Stable lowercase names for severity levels.
const (
	severityNameInfo     = "info"
	severityNameLow      = "low"
	severityNameMedium   = "medium"
	severityNameHigh     = "high"
	severityNameCritical = "critical"
	severityNameUnknown  = "unknown"
)

// Stable lowercase names for exposure levels.
const (
	exposureNameLoopback      = "loopback"
	exposureNamePrivate       = "private"
	exposureNameAllInterfaces = "all-interfaces"
	exposureNamePublic        = "public"
	exposureNameUnknown       = "unknown"
)

// Supported transport protocols.
const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

// Known port sources.
const (
	SourceHost     Source = "host"
	SourceDocker   Source = "docker"
	SourceExternal Source = "external"
)

// Severity levels.
const (
	Info Severity = iota
	Low
	Medium
	High
	Critical
)

// Exposure levels.
const (
	Loopback Exposure = iota
	Private
	AllInterfaces
	Public
)
