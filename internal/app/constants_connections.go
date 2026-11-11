package app

// maxTopRemoteAddresses bounds the top remote address list of a summary.
const maxTopRemoteAddresses = 10

// Error message formats.
const (
	noConnectionSourceMessage = "app: no connection source configured"
	connectionSourceFormat    = "app: list connections: %w"
)
