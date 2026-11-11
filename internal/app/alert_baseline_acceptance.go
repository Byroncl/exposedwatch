package app

import (
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// BaselineAcceptance answers whether a finding's port is part of the accepted
// baseline. A port is accepted only when its canonical identity (protocol,
// port number, bind address and source) is listed; nothing else ever matches.
type BaselineAcceptance struct {
	accepted map[domain.BaselineEntry]bool
	savedOn  string
}

// NewBaselineAcceptance builds the lookup from a trusted baseline.
func NewBaselineAcceptance(baseline domain.Baseline) *BaselineAcceptance {
	acceptance := &BaselineAcceptance{
		accepted: make(map[domain.BaselineEntry]bool, len(baseline.Entries)),
		savedOn:  baseline.CreatedAt.UTC().Format(baselineDateLayout),
	}
	for _, entry := range baseline.Entries {
		acceptance.accepted[entry.Canonical()] = true
	}
	return acceptance
}

// Accepts reports whether the port is in the baseline.
func (acceptance *BaselineAcceptance) Accepts(port domain.Port) bool {
	return acceptance.accepted[domain.BaselineEntryOf(port)]
}

// NotAcceptedLine is the alert body line for a finding outside the baseline.
func (acceptance *BaselineAcceptance) NotAcceptedLine() string {
	return domain.SanitizeLine(fmt.Sprintf(notInBaselineFormat, acceptance.savedOn))
}

// EntryCount is the number of distinct accepted ports.
func (acceptance *BaselineAcceptance) EntryCount() int {
	return len(acceptance.accepted)
}

// SavedOn is the date the baseline was saved (UTC, yyyy-mm-dd).
func (acceptance *BaselineAcceptance) SavedOn() string {
	return acceptance.savedOn
}
