package tui

import "github.com/Byroncl/exposedwatch/internal/app"

// buildRows converts the current data into unfiltered rows for a tab.
func (model Model) buildRows(tab TabID) []displayRow {
	switch tab {
	case TabFindings:
		return model.findingRows()
	case TabAllPorts:
		return model.portRows()
	case TabConnections:
		return model.connectionRows()
	case TabFirewallPorts:
		return model.firewallRows()
	case TabChanges:
		return model.changeRows()
	default:
		return nil
	}
}

func (model Model) decorate(row displayRow) displayRow {
	identity := app.IdentityOf(row.port)
	row.isNew = model.tracker.isNew(identity)
	row.firstSeen, row.hasFirstSeen = model.tracker.firstSeenAt(identity)
	return model.withFirewall(row)
}

func (model Model) findingRows() []displayRow {
	rows := make([]displayRow, 0, len(model.report.Findings))
	for _, finding := range model.report.Findings {
		row := displayRow{
			key:            rowKey{ruleID: finding.RuleID, identity: app.IdentityOf(finding.Port), container: finding.Port.Container},
			port:           finding.Port,
			severity:       finding.Severity,
			hasSeverity:    true,
			title:          finding.Title,
			ruleID:         finding.RuleID,
			recommendation: finding.Recommendation,
		}
		rows = append(rows, model.decorate(row))
	}
	return rows
}

func (model Model) portRows() []displayRow {
	strongest := make(map[app.PortIdentity]int, len(model.report.Findings))
	for index, finding := range model.report.Findings {
		identity := app.IdentityOf(finding.Port)
		existing, known := strongest[identity]
		if !known || finding.Severity > model.report.Findings[existing].Severity {
			strongest[identity] = index
		}
	}
	rows := make([]displayRow, 0, len(model.report.Ports))
	for _, port := range model.report.Ports {
		identity := app.IdentityOf(port)
		row := displayRow{key: rowKey{identity: identity, container: port.Container}, port: port}
		if index, found := strongest[identity]; found {
			finding := model.report.Findings[index]
			row.severity, row.hasSeverity = finding.Severity, true
			row.title, row.ruleID, row.recommendation = finding.Title, finding.RuleID, finding.Recommendation
		}
		rows = append(rows, model.decorate(row))
	}
	return rows
}

// changeRows lists the change log newest first.
func (model Model) changeRows() []displayRow {
	log := model.tracker.log
	rows := make([]displayRow, 0, len(log))
	for index := len(log) - 1; index >= 0; index-- {
		entry := log[index]
		row := displayRow{
			key: rowKey{
				identity: app.IdentityOf(entry.Change.Port), container: entry.Change.Port.Container,
				kind: entry.Change.Kind, changedAt: entry.At,
			},
			port:        entry.Change.Port,
			severity:    entry.Severity,
			hasSeverity: entry.HasSeverity,
			isChange:    true,
			changeKind:  entry.Change.Kind,
			changedAt:   entry.At,
		}
		rows = append(rows, row)
	}
	return rows
}
