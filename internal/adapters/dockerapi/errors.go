package dockerapi

import "errors"

// ErrDockerUnavailable means the Docker socket could not be reached.
var ErrDockerUnavailable = errors.New(unavailableText)
