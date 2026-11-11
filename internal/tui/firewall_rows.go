package tui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// firewallPortKey identifies the port a firewall finding belongs to.
type firewallPortKey struct {
	identity  app.PortIdentity
	container string
}

func firewallKeyOf(port domain.Port) firewallPortKey {
	return firewallPortKey{identity: app.IdentityOf(port), container: port.Container}
}

// indexFirewallFindings maps every evaluated port to its finding.
func indexFirewallFindings(findings []domain.FirewallFinding) map[firewallPortKey]domain.FirewallFinding {
	index := make(map[firewallPortKey]domain.FirewallFinding, len(findings))
	for _, finding := range findings {
		index[firewallKeyOf(finding.Port)] = finding
	}
	return index
}

// withFirewall attaches the verdict of the port, when it was evaluated.
func (model Model) withFirewall(row displayRow) displayRow {
	finding, evaluated := model.firewallIndex[firewallKeyOf(row.port)]
	row.firewall, row.hasFirewall = finding, evaluated
	return row
}

// firewallRows lists one unsorted row per evaluated port.
func (model Model) firewallRows() []displayRow {
	rows := make([]displayRow, 0, len(model.firewallReport.Findings))
	for _, finding := range model.firewallReport.Findings {
		rows = append(rows, displayRow{
			key:         rowKey{identity: app.IdentityOf(finding.Port), container: finding.Port.Container},
			port:        finding.Port,
			firewall:    finding,
			hasFirewall: true,
			isFirewall:  true,
		})
	}
	return rows
}

// matchesFirewallText reports whether any displayed field contains the text.
func (row displayRow) matchesFirewallText(text string) bool {
	if text == "" {
		return true
	}
	port := row.port
	fields := []string{
		strconv.Itoa(port.Number), string(port.Protocol), port.BindAddress, port.Process, port.Container,
		firewallWord(row.firewall.Verdict), row.firewall.Reason,
	}
	return strings.Contains(strings.ToLower(strings.Join(fields, " ")), strings.ToLower(text))
}

// arrangeFirewallRows filters the rows with the text and sorts them.
func (model Model) arrangeFirewallRows(rows []displayRow) []displayRow {
	filtered := make([]displayRow, 0, len(rows))
	for _, row := range rows {
		if row.matchesFirewallText(model.firewallFilterText) {
			filtered = append(filtered, row)
		}
	}
	sort.SliceStable(filtered, func(left, right int) bool {
		return firewallRowLess(filtered[left], filtered[right], model.firewallSortMode)
	})
	return filtered
}

func firewallRowLess(left, right displayRow, mode FirewallSortMode) bool {
	switch mode {
	case FirewallSortByPort:
		return lessByPortThenVerdict(left, right)
	case FirewallSortByProcess:
		leftName := strings.ToLower(processOrContainer(left.port))
		rightName := strings.ToLower(processOrContainer(right.port))
		if leftName != rightName {
			return leftName < rightName
		}
		return lessByPortThenVerdict(left, right)
	default:
		leftRank, rightRank := firewallRank(left.firewall.Verdict), firewallRank(right.firewall.Verdict)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return lessByPortThenVerdict(left, right)
	}
}

func lessByPortThenVerdict(left, right displayRow) bool {
	if left.port.Number != right.port.Number {
		return left.port.Number < right.port.Number
	}
	return firewallRank(left.firewall.Verdict) < firewallRank(right.firewall.Verdict)
}
