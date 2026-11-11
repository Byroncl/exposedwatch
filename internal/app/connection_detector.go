package app

import (
	"sort"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ConnectionChange is one connection that opened or closed between two
// snapshots.
type ConnectionChange struct {
	Kind       ChangeKind
	Connection domain.Connection
}

// DetectConnectionChanges compares two snapshots and lists the connections
// that opened and closed. A nil previous snapshot means "no baseline yet" and
// yields no changes; a non-nil empty snapshot is a real baseline. Duplicates
// collapse to one connection. The result is ordered: opened first, then
// closed, each by connection identity.
func DetectConnectionChanges(previous, current []domain.Connection) []ConnectionChange {
	if previous == nil {
		return nil
	}
	previousByIdentity := indexConnectionsByIdentity(previous)
	currentByIdentity := indexConnectionsByIdentity(current)
	var changes []ConnectionChange
	for identity, connection := range currentByIdentity {
		if _, existed := previousByIdentity[identity]; !existed {
			changes = append(changes, ConnectionChange{Kind: ChangeOpened, Connection: connection})
		}
	}
	for identity, connection := range previousByIdentity {
		if _, stillOpen := currentByIdentity[identity]; !stillOpen {
			changes = append(changes, ConnectionChange{Kind: ChangeClosed, Connection: connection})
		}
	}
	sort.Slice(changes, func(left, right int) bool {
		if changes[left].Kind != changes[right].Kind {
			return changes[left].Kind < changes[right].Kind
		}
		return ConnectionIdentityLess(
			ConnectionIdentityOf(changes[left].Connection),
			ConnectionIdentityOf(changes[right].Connection),
		)
	})
	return changes
}

// indexConnectionsByIdentity keeps the first connection seen per identity.
func indexConnectionsByIdentity(connections []domain.Connection) map[ConnectionIdentity]domain.Connection {
	indexed := make(map[ConnectionIdentity]domain.Connection, len(connections))
	for _, connection := range connections {
		identity := ConnectionIdentityOf(connection)
		if _, duplicate := indexed[identity]; !duplicate {
			indexed[identity] = connection
		}
	}
	return indexed
}
