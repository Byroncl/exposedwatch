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

func TestParseSeverity(testRun *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    Severity
		wantErr bool
	}{
		{"info", "info", Info, false},
		{"low", "low", Low, false},
		{"medium", "medium", Medium, false},
		{"high", "high", High, false},
		{"critical", "critical", Critical, false},
		{"upper case", "HIGH", High, false},
		{"mixed case", "MeDiUm", Medium, false},
		{"surrounding spaces", "  low ", Low, false},
		{"unknown", "severe", Info, true},
		{"empty", "", Info, true},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(subRun *testing.T) {
			got, err := ParseSeverity(testCase.input)
			if (err != nil) != testCase.wantErr {
				subRun.Fatalf("ParseSeverity(%q) error = %v, wantErr %v", testCase.input, err, testCase.wantErr)
			}
			if err == nil && got != testCase.want {
				subRun.Fatalf("ParseSeverity(%q) = %v, want %v", testCase.input, got, testCase.want)
			}
		})
	}
}
