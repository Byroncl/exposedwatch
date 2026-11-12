package domain

// Alert kinds.
const (
	AlertKindNewFinding AlertKind = "new-finding"
	AlertKindSummary    AlertKind = "summary"
	AlertKindTest       AlertKind = "test"
)

// Text used when an alert credential is printed by accident.
const redactedCredentialText = "***"

// Invisible formatting characters removed from untrusted text.
const (
	bidiEmbeddingStart rune = 0x202A
	bidiEmbeddingEnd   rune = 0x202E
	bidiIsolateStart   rune = 0x2066
	bidiIsolateEnd     rune = 0x2069
	zeroWidthStart     rune = 0x200B
	zeroWidthEnd       rune = 0x200D
	byteOrderMark      rune = 0xFEFF
)

// portLabelFormat renders "port/protocol", for example "5432/tcp".
const portLabelFormat = "%d/%s"
