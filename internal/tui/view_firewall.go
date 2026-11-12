package tui

import (
	"fmt"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// firewallStatusPanel renders the Status inner tab of the Firewall module.
func (model Model) firewallStatusPanel(frame layout) []string {
	innerWidth := panelContentWidth(frame.contentWidth)
	lines := model.firewallStatusLines(innerWidth)
	lines = firstLines(lines, max(frame.panelLines-panelBorderLineCount, 0))
	lines = append(lines, blankLines(frame.panelLines-panelBorderLineCount-len(lines), innerWidth)...)
	return model.styles.renderPanel(panelFirewallTitle, lines, frame.contentWidth)
}

func (model Model) firewallStatusLines(innerWidth int) []string {
	switch {
	case !model.firewallEnabled():
		return wrapDim(model.styles, firewallDisabledText, innerWidth)
	case !model.hasFirewall && model.firewallErr != nil:
		return wrapDim(model.styles, firewallFailedText, innerWidth)
	case !model.hasFirewall:
		return wrapDim(model.styles, firewallReadingText, innerWidth)
	}
	lines := []string{
		model.styles.heading.Render(fmt.Sprintf(firewallBackendFormat, model.firewallReport.Backend, availableText(model.firewallReport.Available))),
		model.styles.dim.Render(fmt.Sprintf(firewallFreshFormat, model.firewallAt.Format(timeLayout), firewallRefreshInterval)),
		"",
		model.firewallChipsLine(),
	}
	lines = append(lines, model.firewallNoteLines(innerWidth)...)
	lines = append(lines, "", model.styles.heading.Render(firewallLegendHeading))
	lines = append(lines, model.firewallLegendLines(innerWidth)...)
	lines = append(lines, "")
	return append(lines, wrapDim(model.styles, firewallInformational+" "+firewallEstimateNote, innerWidth)...)
}

func availableText(available bool) string {
	if available {
		return firewallAvailableYes
	}
	return firewallAvailableNo
}

// wrapDim wraps plain text to the width and dims every line.
func wrapDim(palette styles, text string, width int) []string {
	wrapped := wrapText(text, width)
	for index, line := range wrapped {
		wrapped[index] = palette.dim.Render(line)
	}
	return wrapped
}

// firewallChipsLine shows one chip per verdict with its count; empty ones are dim.
func (model Model) firewallChipsLine() string {
	counts := countFirewallVerdicts(model.firewallReport.Findings)
	parts := make([]string, 0, len(firewallVerdictOrder)+1)
	parts = append(parts, model.styles.dim.Render(firewallVerdictsLabel))
	for _, verdict := range firewallVerdictOrder {
		label := " " + strings.ToUpper(fmt.Sprintf(firewallChipCountFormat, firewallWord(verdict), counts[verdict])) + " "
		if counts[verdict] == 0 {
			parts = append(parts, model.styles.dim.Render(label))
			continue
		}
		parts = append(parts, model.styles.chipForVerdict(verdict).Render(label))
	}
	return strings.Join(parts, "")
}

// firewallNoteLines lists the notes of the engine under a heading, if any.
func (model Model) firewallNoteLines(innerWidth int) []string {
	if len(model.firewallReport.Notes) == 0 {
		return nil
	}
	lines := []string{"", model.styles.heading.Render(firewallNotesHeading)}
	for _, note := range model.firewallReport.Notes {
		lines = append(lines, wrapText(fmt.Sprintf(firewallNoteFormat, note), innerWidth)...)
	}
	return lines
}

// firewallLegendLines explains what each verdict means, one entry per verdict.
func (model Model) firewallLegendLines(innerWidth int) []string {
	meanings := map[domain.FirewallVerdict]string{
		domain.FirewallAllowed:         legendAllowed,
		domain.FirewallRestricted:      legendRestricted,
		domain.FirewallBlocked:         legendBlocked,
		domain.FirewallUnknown:         legendUnknown,
		domain.FirewallNotExposed:      legendNotExposed,
		domain.FirewallDockerPublished: legendDocker,
	}
	var lines []string
	for _, verdict := range firewallVerdictOrder {
		entry := firewallWord(verdict) + firewallLegendSeparator + meanings[verdict]
		lines = append(lines, wrapText(entry, innerWidth)...)
	}
	return lines
}

// overviewFirewallLine shows the firewall counters of the Overview, or why there are none.
func (model Model) overviewFirewallLine() string {
	switch {
	case !model.firewallEnabled():
		return columnSeparator + model.styles.dim.Render(firewallOverviewDisabled)
	case !model.hasFirewall:
		return columnSeparator + model.styles.dim.Render(firewallOverviewReading)
	case !model.firewallReport.Available:
		reason := firewallOverviewNoReason
		if len(model.firewallReport.Notes) > 0 {
			reason = model.firewallReport.Notes[0]
		}
		return columnSeparator + model.styles.warning.Render(fmt.Sprintf(firewallOverviewMissing, reason))
	}
	counts := countFirewallVerdicts(model.firewallReport.Findings)
	return columnSeparator + model.styles.plain.Render(fmt.Sprintf(firewallOverviewFormat,
		counts[domain.FirewallAllowed], counts[domain.FirewallRestricted], counts[domain.FirewallBlocked], counts[domain.FirewallUnknown]))
}
