package tui

import (
	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// RiskLevel is the overall risk of a report, driven by its highest severity.
type RiskLevel int

// Risk levels, from calmest to worst.
const (
	RiskSafe RiskLevel = iota
	RiskLow
	RiskMedium
	RiskHigh
	RiskCritical
)

// riskLevelOf maps the highest severity present to a level. Info findings and
// an empty report are both SAFE.
func riskLevelOf(report app.Report) RiskLevel {
	switch report.HighestSeverity() {
	case domain.Critical:
		return RiskCritical
	case domain.High:
		return RiskHigh
	case domain.Medium:
		return RiskMedium
	case domain.Low:
		return RiskLow
	default:
		return RiskSafe
	}
}

// String returns the uppercase word shown next to the gauge.
func (level RiskLevel) String() string {
	switch level {
	case RiskLow:
		return riskWordLow
	case RiskMedium:
		return riskWordMedium
	case RiskHigh:
		return riskWordHigh
	case RiskCritical:
		return riskWordCritical
	default:
		return riskWordSafe
	}
}

// gaugeCells returns how many gauge cells are filled and empty.
func (level RiskLevel) gaugeCells() (filled, empty int) {
	filled = (int(level) + 1) * riskGaugeWidth / (int(RiskCritical) + 1)
	return filled, riskGaugeWidth - filled
}

// severity returns the severity whose color paints this level.
func (level RiskLevel) severity() (domain.Severity, bool) {
	switch level {
	case RiskLow:
		return domain.Low, true
	case RiskMedium:
		return domain.Medium, true
	case RiskHigh:
		return domain.High, true
	case RiskCritical:
		return domain.Critical, true
	default:
		return domain.Info, false
	}
}
