package nmapscan

import "errors"

// ErrNmapUnavailable means the nmap executable could not be found or started.
var ErrNmapUnavailable = errors.New("nmapscan: nmap is not installed or not executable")

// ErrNmapFailed means nmap ran and exited with a non-zero status.
var ErrNmapFailed = errors.New("nmapscan: nmap failed")

// ErrInvalidOutput means nmap's XML output could not be understood.
var ErrInvalidOutput = errors.New("nmapscan: unparseable nmap output")

// ErrInvalidTarget means a scan target failed validation.
var ErrInvalidTarget = errors.New("nmapscan: invalid target")

// ErrNoTargets means there is nothing to scan.
var ErrNoTargets = errors.New("nmapscan: no targets to scan")

// ErrInvalidPortSpec means the port list failed validation.
var ErrInvalidPortSpec = errors.New("nmapscan: invalid port list")

// ErrInvalidTimeout means the scan timeout is out of bounds.
var ErrInvalidTimeout = errors.New("nmapscan: invalid timeout")

// ErrScanTimedOut means the scan exceeded its time limit and nmap was killed.
var ErrScanTimedOut = errors.New("nmapscan: scan timed out")
