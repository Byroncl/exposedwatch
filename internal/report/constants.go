package report

// ColorMode says whether ANSI colors are emitted.
type ColorMode bool

// Color modes.
const (
	ColorEnabled  ColorMode = true
	ColorDisabled ColorMode = false
)

// AnsiCode is a terminal escape sequence.
type AnsiCode string

// ANSI escape sequences.
const (
	ansiReset            AnsiCode = "\x1b[0m"
	ansiBold             AnsiCode = "\x1b[1m"
	ansiGray             AnsiCode = "\x1b[90m"
	ansiCyan             AnsiCode = "\x1b[36m"
	ansiYellow           AnsiCode = "\x1b[33m"
	ansiRed              AnsiCode = "\x1b[31m"
	ansiBoldRedOnDefault AnsiCode = "\x1b[1;91m"
)

// Environment and terminal detection.
const (
	// noColorDisablesWhenSet follows https://no-color.org: any non-empty value disables color.
	noColorUnsetValue = ""
)

// Table column headers.
const (
	headerSeverity = "SEVERITY"
	headerPort     = "PORT"
	headerProtocol = "PROTOCOL"
	headerBind     = "BIND ADDRESS"
	headerExposure = "EXPOSURE"
	headerSource   = "SOURCE"
	headerProcess  = "PROCESS/CONTAINER"
	headerTitle    = "TITLE"
)

// Position of the FIREWALL column when it is shown: right before the process.
const firewallColumnIndex = 6

// Table layout.
const (
	columnSeparator  = "  "
	emptyCellText    = "-"
	paddingCharacter = " "
)

// Table texts.
const (
	noFindingsText       = "No findings: nothing is exposed beyond loopback."
	recommendationsTitle = "Recommendations"
	recommendationFormat = "  - [%s] %s\n"
	failedSourcesTitle   = "Sources that could not be read:"
	failedSourceFormat   = "  - %s\n"
	summaryFormat        = "%d port(s) scanned, %d finding(s).\n"
	newline              = "\n"
	jsonIndent           = "  "
	jsonPrefix           = ""
)
