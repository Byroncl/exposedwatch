package tui

import "fmt"

// firewallDetailLines describes the selected row of the Firewall ports table,
// with the full reason of the verdict.
func (model Model) firewallDetailLines(row displayRow, contentWidth int) []string {
	verdict := row.firewall.Verdict
	headline := model.styles.chipForVerdict(verdict).Render(columnSeparator+firewallWord(verdict)+columnSeparator) +
		columnSeparator + fmt.Sprintf(portFormat, row.port.Number, row.port.Protocol)
	lines := []string{
		headline,
		model.bindDetail(row),
		model.processDetail(row),
		fmt.Sprintf(detailSourceFormat, row.port.Source),
	}
	return append(lines, wrapText(fmt.Sprintf(detailReasonFormat, row.firewall.Reason), contentWidth)...)
}

// firewallDetailLine is the Firewall line of the port details: the verdict and
// its reason, or a dash when the port was not evaluated.
func firewallDetailLine(row displayRow) string {
	if !row.hasFirewall {
		return fmt.Sprintf(detailFirewallFormat, emptyValue)
	}
	return fmt.Sprintf(detailFirewallReasonFormat, firewallWord(row.firewall.Verdict), row.firewall.Reason)
}
