package iptablessave

import "errors"

// ErrInvalidOutput means the output of a save command could not be understood.
var ErrInvalidOutput = errors.New("iptablessave: unparseable iptables-save output")

// ErrCommandFailed means a save command ran and exited with a non-zero status.
var ErrCommandFailed = errors.New("iptablessave: save command failed")

// ErrBinaryNotFound means a save executable is not installed.
var ErrBinaryNotFound = errors.New("iptablessave: save executable not found")
