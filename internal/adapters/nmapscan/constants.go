package nmapscan

import "time"

// BinaryName is the default nmap executable name, resolved through PATH.
const BinaryName = "nmap"

// Timeout bounds accepted for a scan.
const (
	DefaultTimeout = 2 * time.Minute
	MinTimeout     = 5 * time.Second
	MaxTimeout     = 30 * time.Minute

	// processGrace is added to the scan timeout so nmap's own
	// --host-timeout fires before the process is killed.
	processGrace = 10 * time.Second

	// processWaitDelay bounds how long Wait blocks on output pipes after the
	// process is killed, in case a grandchild keeps them open.
	processWaitDelay = 3 * time.Second
)

// Fixed nmap command line. Targets always follow terminatorArgument so they
// can never be read as options.
const (
	outputFlag         = "-oX"
	outputToStdout     = "-"
	skipHostDiscovery  = "-Pn"
	skipDNSResolution  = "-n"
	onlyOpenPorts      = "--open"
	tcpConnectScan     = "-sT"
	timingFlag         = "-T3"
	hostTimeoutFlag    = "--host-timeout"
	topPortsFlag       = "--top-ports"
	portsFlag          = "-p"
	terminatorArgument = "--"
	defaultTopPorts    = "1000"
	secondsFormat      = "%ds"
)

// Output limits.
const (
	maxStdoutBytes      = 16 * 1024 * 1024
	maxStderrBytes      = 64 * 1024
	maxStderrCharacters = 512
)

// XML values read from nmap output.
const (
	xmlRootElement   = "nmaprun"
	stateOpen        = "open"
	protocolTCP      = "tcp"
	addressTypeIPv4  = "ipv4"
	addressTypeIPv6  = "ipv6"
	noExitCode       = 0
	killedExitStatus = -1
)

// Target validation limits.
const (
	maxTargetCount      = 64
	maxHostnameLength   = 253
	maxLabelLength      = 63
	minIPv4PrefixBits   = 24
	hostnameSeparator   = "."
	labelHyphen         = '-'
	maxPortNumber       = 65535
	minPortNumber       = 1
	maxPortSpecLength   = 256
	portListSeparator   = ","
	portRangeSeparator  = "-"
	portRangePartsCount = 2
)

// Error message formats.
const (
	invalidTargetFormat   = "%w: %q: %s"
	invalidPortSpecFormat = "%w: %q: %s"
	nmapFailedFormat      = "%w: exit code %d: %s"
	invalidOutputFormat   = "%w: %v"
	timeoutFormat         = "%w: %v"
	invalidTimeoutFormat  = "%w: %v, want between %v and %v"
	startFailedFormat     = "nmapscan: start nmap: %w"
	unavailableFormat     = "%w: %v"
	listAddressesFormat   = "nmapscan: list interface addresses: %w"
	tooManyTargetsFormat  = "more than %d targets"
)

// Reasons attached to validation errors.
const (
	reasonEmpty            = "empty"
	reasonTooLong          = "too long"
	reasonLeadingHyphen    = "must not start with '-'"
	reasonBadCharacters    = "contains characters outside letters, digits, dots and hyphens"
	reasonBadLabel         = "invalid hostname label"
	reasonNumericHostname  = "hostname must end with a label that starts with a letter"
	reasonIPv6NotSupported = "IPv6 targets are not supported yet, use IPv4 addresses or hostnames"
	reasonZoneNotAllowed   = "IPv6 zones are not allowed"
	reasonPrefixTooWide    = "network is wider than the allowed prefix"
	reasonBadPortItem      = "invalid port or range"
	reasonPortOutOfRange   = "port out of range 1-65535"
	reasonRangeOrder       = "range start is greater than its end"
	reasonPortSpecCharsets = "only digits, commas and hyphens are allowed"
)
