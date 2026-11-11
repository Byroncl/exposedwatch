package risk

import (
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func port(n int, addr string) domain.Port {
	return domain.Port{Number: n, Protocol: domain.TCP, BindAddress: addr, Source: domain.SourceHost}
}

func TestEvaluateRules(t *testing.T) {
	tests := []struct {
		name     string
		port     domain.Port
		wantNone bool
		want     domain.Severity
	}{
		{"ssh all interfaces", port(22, "0.0.0.0"), false, domain.Medium},
		{"telnet all interfaces", port(23, "::"), false, domain.Critical},
		{"mysql all", port(3306, "0.0.0.0"), false, domain.High},
		{"postgres all", port(5432, ""), false, domain.High},
		{"redis all", port(6379, "0.0.0.0"), false, domain.High},
		{"mongo all", port(27017, "0.0.0.0"), false, domain.High},
		{"elasticsearch all", port(9200, "0.0.0.0"), false, domain.High},
		{"memcached all", port(11211, "0.0.0.0"), false, domain.High},
		{"redis private lan", port(6379, "192.168.1.5"), false, domain.High},
		{"docker api all", port(2375, "0.0.0.0"), false, domain.Critical},
		{"rdp all", port(3389, "0.0.0.0"), false, domain.High},
		{"smb 445", port(445, "0.0.0.0"), false, domain.High},
		{"smb 139", port(139, "0.0.0.0"), false, domain.High},
		{"ssh public ip", port(22, "203.0.113.5"), false, domain.Medium},
		{"ssh private only", port(22, "10.0.0.2"), true, 0},
		{"generic all interfaces", port(8080, "0.0.0.0"), false, domain.Info},
		{"generic private", port(8080, "10.0.0.2"), true, 0},
		{"redis loopback", port(6379, "127.0.0.1"), true, 0},
		{"docker api loopback ipv6", port(2375, "::1"), true, 0},
		{"generic loopback", port(8080, "127.0.0.1"), true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate([]domain.Port{tt.port})
			if tt.wantNone {
				if len(got) != 0 {
					t.Fatalf("expected no findings, got %+v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("expected 1 finding, got %+v", got)
			}
			f := got[0]
			if f.Severity != tt.want {
				t.Errorf("severity = %v, want %v", f.Severity, tt.want)
			}
			if f.RuleID == "" || f.Title == "" || f.Recommendation == "" {
				t.Errorf("incomplete finding: %+v", f)
			}
			if f.Port.Number != tt.port.Number {
				t.Errorf("finding port = %d, want %d", f.Port.Number, tt.port.Number)
			}
		})
	}
}

func TestEvaluateEmpty(t *testing.T) {
	if got := Evaluate(nil); len(got) != 0 {
		t.Fatalf("nil input: got %+v", got)
	}
	if got := Evaluate([]domain.Port{}); len(got) != 0 {
		t.Fatalf("empty input: got %+v", got)
	}
}

func TestEvaluateSorting(t *testing.T) {
	in := []domain.Port{
		port(9000, "0.0.0.0"), // info
		port(6379, "0.0.0.0"), // high
		port(23, "0.0.0.0"),   // critical
		port(3306, "0.0.0.0"), // high
		port(8080, "0.0.0.0"), // info
		port(22, "127.0.0.1"), // none
		port(22, "0.0.0.0"),   // medium
	}
	got := Evaluate(in)
	wantPorts := []int{23, 3306, 6379, 22, 8080, 9000}
	if len(got) != len(wantPorts) {
		t.Fatalf("got %d findings, want %d: %+v", len(got), len(wantPorts), got)
	}
	for i, n := range wantPorts {
		if got[i].Port.Number != n {
			t.Errorf("position %d: port %d, want %d", i, got[i].Port.Number, n)
		}
	}
}
