package tui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// firewallEnabled tells whether a firewall evaluator was configured.
func (model Model) firewallEnabled() bool {
	return model.firewall != nil
}

// startFirewallRead starts an evaluation of the current host ports when one is
// due and returns its command, or nil. It needs a scan report and never runs
// while another evaluation is in flight. An evaluation is due when none ran
// yet, when the listening ports changed, when the throttle window passed or
// when force is set (the refresh key).
func (model *Model) startFirewallRead(force bool) tea.Cmd {
	if !model.firewallEnabled() || model.firewallRefreshing || !model.hasReport {
		return nil
	}
	now := model.clock()
	signature := firewallPortSignature(model.report.Ports)
	due := force || model.firewallRequestedAt.IsZero() || signature != model.firewallSignature ||
		now.Sub(model.firewallRequestedAt) >= firewallRefreshInterval
	if !due {
		return nil
	}
	model.firewallRefreshing = true
	model.firewallRequestedAt = now
	model.firewallSignature = signature
	return firewallCommand(model.firewall, slices.Clone(model.report.Ports), model.clock)
}

// firewallPortSignature identifies the set of host and Docker ports that the
// firewall evaluation covers, ignoring order and ports seen from outside.
func firewallPortSignature(listening []domain.Port) string {
	entries := make([]string, 0, len(listening))
	for _, port := range listening {
		if port.Source == domain.SourceExternal {
			continue
		}
		entries = append(entries, strings.Join([]string{
			strconv.Itoa(port.Number), string(port.Protocol), port.BindAddress, string(port.Source), port.Container,
		}, firewallSignatureJoin))
	}
	slices.Sort(entries)
	return strings.Join(entries, firewallEntryJoin)
}

// handleFirewallResult stores an evaluation. A failed one keeps the previous
// data and only raises the warning of the Firewall views.
func (model Model) handleFirewallResult(result FirewallResultMessage) Model {
	model.firewallRefreshing = false
	model.firewallErr = result.Err
	if result.Err != nil {
		return model
	}
	model.firewallReport = result.Report
	model.firewallIndex = indexFirewallFindings(result.Report.Findings)
	model.hasFirewall = true
	model.firewallAt = result.At
	model.rebuildVisibleRows()
	return model
}

// firewallWarningLines lists the non-fatal firewall problem. It only shows on
// the Firewall views, where it matters.
func (model Model) firewallWarningLines() []string {
	if !model.activeTab.usesFirewall() || !model.firewallEnabled() || model.firewallErr == nil {
		return nil
	}
	return []string{fmt.Sprintf(firewallWarningFormat, model.firewallErr)}
}

// firewallEmptyMessage explains why the Firewall ports table has no rows.
func (model Model) firewallEmptyMessage() string {
	switch {
	case !model.firewallEnabled():
		return firewallDisabledText
	case !model.hasFirewall && model.firewallErr != nil:
		return firewallFailedText
	case !model.hasFirewall:
		return firewallReadingText
	case model.firewallFilterText != "":
		return noMatchMessage
	default:
		return noFirewallPortsText
	}
}

// firewallBadge is the sidebar note of the Firewall module: off when disabled,
// n/a when the rules could not be read, otherwise the number of ports that
// are allowed or restricted.
func (model Model) firewallBadge() sidebarBadge {
	switch {
	case !model.firewallEnabled():
		return sidebarBadge{text: sidebarBadgeOff}
	case !model.hasFirewall:
		return sidebarBadge{}
	case !model.firewallReport.Available:
		return sidebarBadge{text: sidebarBadgeNotAvailable}
	}
	counts := countFirewallVerdicts(model.firewallReport.Findings)
	return sidebarBadge{text: cappedCount(counts[domain.FirewallAllowed] + counts[domain.FirewallRestricted])}
}

// firewallSettingText says whether the firewall checks run, for the Settings placeholder.
func (model Model) firewallSettingText() string {
	if model.firewallEnabled() {
		return settingsFirewallOn
	}
	return settingsFirewallOff
}
