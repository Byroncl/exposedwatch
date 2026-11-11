package app

import "errors"

// ErrNoSources is returned when a scan has no port sources to query.
var ErrNoSources = errors.New(noSourcesMessage)

// ErrAllSourcesFailed is returned when every configured source failed.
var ErrAllSourcesFailed = errors.New(allSourcesFailedMessage)
