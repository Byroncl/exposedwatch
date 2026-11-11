package app

import "github.com/Byroncl/exposedwatch/internal/domain"

// ChangeKind tells whether a port appeared or disappeared between snapshots.
type ChangeKind int

// Change kinds.
const (
	ChangeOpened ChangeKind = iota
	ChangeClosed
)

// String returns a stable lowercase name.
func (kind ChangeKind) String() string {
	switch kind {
	case ChangeOpened:
		return changeKindNameOpened
	case ChangeClosed:
		return changeKindNameClosed
	default:
		return changeKindNameUnknown
	}
}

// Change is one port that opened or closed between two snapshots.
type Change struct {
	Kind ChangeKind
	Port domain.Port
}
