package dockerapi

// SocketPath is a filesystem path of a unix socket.
type SocketPath string

// DefaultSocketPath is the standard Docker Engine socket path.
const DefaultSocketPath SocketPath = "/var/run/docker.sock"

// Transport and endpoint parameters.
const (
	unixNetwork           = "unix"
	dialOperation         = "dial"
	containersURL         = "http://docker/containers/json"
	containerNamePrefix   = "/"
	unpublishedPortNumber = 0
)

// Error message formats.
const (
	unavailableMessage    = "%w: %v"
	requestFailedMessage  = "dockerapi: request failed: %w"
	unexpectedStatusFmt   = "dockerapi: unexpected status %s"
	decodeResponseMessage = "dockerapi: decode response: %w"
	unavailableText       = "dockerapi: docker engine unavailable"
)
