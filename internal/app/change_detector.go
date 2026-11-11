package app

import (
	"sort"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// DetectChanges compares two snapshots and lists the ports that opened and
// closed. A nil previous snapshot means "no baseline yet" and yields no
// changes; a non-nil empty snapshot is a real baseline with no ports.
// Duplicates collapse to one port. The result is ordered: opened ports first,
// then closed ports, each by identity (number, protocol, bind address, source).
func DetectChanges(previous, current []domain.Port) []Change {
	if previous == nil {
		return nil
	}
	previousByIdentity := indexByIdentity(previous)
	currentByIdentity := indexByIdentity(current)
	var changes []Change
	for identity, port := range currentByIdentity {
		if _, existed := previousByIdentity[identity]; !existed {
			changes = append(changes, Change{Kind: ChangeOpened, Port: port})
		}
	}
	for identity, port := range previousByIdentity {
		if _, stillOpen := currentByIdentity[identity]; !stillOpen {
			changes = append(changes, Change{Kind: ChangeClosed, Port: port})
		}
	}
	sort.Slice(changes, func(left, right int) bool {
		return changeLess(changes[left], changes[right])
	})
	return changes
}

// indexByIdentity keeps the first port seen for each identity.
func indexByIdentity(ports []domain.Port) map[PortIdentity]domain.Port {
	indexed := make(map[PortIdentity]domain.Port, len(ports))
	for _, port := range ports {
		identity := IdentityOf(port)
		if _, duplicate := indexed[identity]; !duplicate {
			indexed[identity] = port
		}
	}
	return indexed
}

func changeLess(left, right Change) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	return IdentityLess(IdentityOf(left.Port), IdentityOf(right.Port))
}

// IdentityLess orders identities by number, protocol, bind address, source.
func IdentityLess(left, right PortIdentity) bool {
	if left.Number != right.Number {
		return left.Number < right.Number
	}
	if left.Protocol != right.Protocol {
		return left.Protocol < right.Protocol
	}
	if left.BindAddress != right.BindAddress {
		return left.BindAddress < right.BindAddress
	}
	return left.Source < right.Source
}
