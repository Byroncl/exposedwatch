package domain

import "testing"

const invalidSeverity Severity = 99

func TestSeverityOrderingAndString(testRun *testing.T) {
	order := []Severity{Info, Low, Medium, High, Critical}
	names := []string{"info", "low", "medium", "high", "critical"}
	for index, severity := range order {
		if severity.String() != names[index] {
			testRun.Errorf("String() = %q, want %q", severity.String(), names[index])
		}
		if index > 0 && order[index-1] >= severity {
			testRun.Errorf("%v should be lower than %v", order[index-1], severity)
		}
	}
	if invalidSeverity.String() != "unknown" {
		testRun.Errorf("unexpected string for invalid severity")
	}
}
