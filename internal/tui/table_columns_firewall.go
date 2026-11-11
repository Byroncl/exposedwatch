package tui

import (
	"strconv"
	"strings"
)

// firewallPortsSpec is the column set of the Firewall ports tab. The verdict,
// the port and the protocol are always shown; the process and the bind
// address appear when there is room, and the reason takes what is left.
func firewallPortsSpec() tableSpec {
	return tableSpec{
		fixed: []column{
			{headingVerdict, verdictColumnWidth, priorityMandatory, chipSidePadding, verdictChipCell},
			{headingPort, portNumberColumnWidth, priorityMandatory, 0, portNumberCell},
			{headingProtocol, protocolColumnWidth, priorityMandatory, 0, protocolCell},
			{headingBind, bindColumnWidth, priorityBind, 0, bindCell},
			{headingProcessShort, processColumnWidth, priorityProcess, 0, processCell},
		},
		flexHeading:  headingReason,
		flexContent:  reasonCell,
		minFlexWidth: minReasonWidth,
	}
}

func verdictChipCell(palette styles, row displayRow) cell {
	word := strings.ToUpper(firewallWord(row.firewall.Verdict))
	return cell{text: word, look: palette.chipForVerdict(row.firewall.Verdict), chip: true}
}

func portNumberCell(palette styles, row displayRow) cell {
	return cell{text: strconv.Itoa(row.port.Number), look: palette.plain}
}

func protocolCell(palette styles, row displayRow) cell {
	return cell{text: string(row.port.Protocol), look: palette.plain}
}

func reasonCell(palette styles, row displayRow) cell {
	return cell{text: row.firewall.Reason, look: palette.plain}
}

// firewallVerdictCell is the low-priority FW column of the port tables: the
// verdict word, or a dash when the port was not evaluated.
func firewallVerdictCell(palette styles, row displayRow) cell {
	if !row.hasFirewall {
		return cell{text: emptyValue, look: palette.dim}
	}
	return cell{text: firewallWord(row.firewall.Verdict), look: palette.plain}
}
