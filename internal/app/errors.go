package app

import "errors"

// ErrNoSources is returned when a scan has no port sources to query.
var ErrNoSources = errors.New(noSourcesMessage)

// ErrAllSourcesFailed is returned when every configured source failed.
var ErrAllSourcesFailed = errors.New(allSourcesFailedMessage)

// ErrNoConnectionSource is returned when a connection service has no source.
var ErrNoConnectionSource = errors.New(noConnectionSourceMessage)
