package tui

import (
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

func TestRiskLevelFollowsHighestSeverity(testRun *testing.T) {
	cases := []struct {
		severities []domain.Severity
		want       RiskLevel
		word       string
	}{
		{nil, RiskSafe, "SAFE"},
		{[]domain.Severity{domain.Info}, RiskSafe, "SAFE"},
		{[]domain.Severity{domain.Info, domain.Low}, RiskLow, "LOW"},
		{[]domain.Severity{domain.Medium, domain.Low}, RiskMedium, "MEDIUM"},
		{[]domain.Severity{domain.High, domain.Medium}, RiskHigh, "HIGH"},
		{[]domain.Severity{domain.Low, domain.Critical}, RiskCritical, "CRITICAL"},
	}
	for _, testCase := range cases {
		var report app.Report
		for _, severity := range testCase.severities {
			report.Findings = append(report.Findings, domain.Finding{Severity: severity})
		}
		got := riskLevelOf(report)
		if got != testCase.want || got.String() != testCase.word {
			testRun.Errorf("%v: level=%v word=%q, want %v %q", testCase.severities, got, got.String(), testCase.want, testCase.word)
		}
	}
}

func TestRiskGaugeHasConstantWidthAndGrowsWithRisk(testRun *testing.T) {
	previousFilled := -1
	for level := RiskSafe; level <= RiskCritical; level++ {
		filled, empty := level.gaugeCells()
		if filled+empty != riskGaugeWidth {
			testRun.Fatalf("level %v: %d+%d != %d", level, filled, empty, riskGaugeWidth)
		}
		if filled <= previousFilled {
			testRun.Fatalf("level %v: filled %d does not grow", level, filled)
		}
		previousFilled = filled
	}
}
