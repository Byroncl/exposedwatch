package risk

import (
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

const (
	portHTTPAlt   = 8080
	portGenericHi = 9000
	addrAll       = "0.0.0.0"
	addrAllIPv6   = "::"
	addrLoopback  = "127.0.0.1"
)

func newPort(number int, address string) domain.Port {
	return domain.Port{Number: number, Protocol: domain.TCP, BindAddress: address, Source: domain.SourceHost}
}

func TestEvaluateRules(testRun *testing.T) {
	testCases := []struct {
		name     string
		port     domain.Port
		wantNone bool
		want     domain.Severity
	}{
		{"ssh all interfaces", newPort(PortSSH, addrAll), false, domain.Medium},
		{"telnet all interfaces", newPort(PortTelnet, addrAllIPv6), false, domain.Critical},
		{"mysql all", newPort(PortMySQL, addrAll), false, domain.High},
		{"postgres all", newPort(PortPostgreSQL, ""), false, domain.High},
		{"redis all", newPort(PortRedis, addrAll), false, domain.High},
		{"mongo all", newPort(PortMongoDB, addrAll), false, domain.High},
		{"elasticsearch all", newPort(PortElasticsearch, addrAll), false, domain.High},
		{"memcached all", newPort(PortMemcached, addrAll), false, domain.High},
		{"redis private lan", newPort(PortRedis, "192.168.1.5"), false, domain.High},
		{"docker api all", newPort(PortDockerAPI, addrAll), false, domain.Critical},
		{"rdp all", newPort(PortRDP, addrAll), false, domain.High},
		{"smb 445", newPort(PortSMB, addrAll), false, domain.High},
		{"smb 139", newPort(PortNetBIOS, addrAll), false, domain.High},
		{"ssh public ip", newPort(PortSSH, "203.0.113.5"), false, domain.Medium},
		{"ssh private only", newPort(PortSSH, "10.0.0.2"), true, 0},
		{"generic all interfaces", newPort(portHTTPAlt, addrAll), false, domain.Info},
		{"generic private", newPort(portHTTPAlt, "10.0.0.2"), true, 0},
		{"redis loopback", newPort(PortRedis, addrLoopback), true, 0},
		{"docker api loopback ipv6", newPort(PortDockerAPI, "::1"), true, 0},
		{"generic loopback", newPort(portHTTPAlt, addrLoopback), true, 0},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(subRun *testing.T) {
			findings := Evaluate([]domain.Port{testCase.port})
			if testCase.wantNone {
				if len(findings) != 0 {
					subRun.Fatalf("expected no findings, got %+v", findings)
				}
				return
			}
			if len(findings) != 1 {
				subRun.Fatalf("expected 1 finding, got %+v", findings)
			}
			finding := findings[0]
			if finding.Severity != testCase.want {
				subRun.Errorf("severity = %v, want %v", finding.Severity, testCase.want)
			}
			if finding.RuleID == "" || finding.Title == "" || finding.Recommendation == "" {
				subRun.Errorf("incomplete finding: %+v", finding)
			}
			if finding.Port.Number != testCase.port.Number {
				subRun.Errorf("finding port = %d, want %d", finding.Port.Number, testCase.port.Number)
			}
		})
	}
}

func TestEvaluateEmpty(testRun *testing.T) {
	if findings := Evaluate(nil); len(findings) != 0 {
		testRun.Fatalf("nil input: got %+v", findings)
	}
	if findings := Evaluate([]domain.Port{}); len(findings) != 0 {
		testRun.Fatalf("empty input: got %+v", findings)
	}
}

func TestEvaluateSorting(testRun *testing.T) {
	input := []domain.Port{
		newPort(portGenericHi, addrAll), // info
		newPort(PortRedis, addrAll),     // high
		newPort(PortTelnet, addrAll),    // critical
		newPort(PortMySQL, addrAll),     // high
		newPort(portHTTPAlt, addrAll),   // info
		newPort(PortSSH, addrLoopback),  // none
		newPort(PortSSH, addrAll),       // medium
	}
	findings := Evaluate(input)
	wantPorts := []int{PortTelnet, PortMySQL, PortRedis, PortSSH, portHTTPAlt, portGenericHi}
	if len(findings) != len(wantPorts) {
		testRun.Fatalf("got %d findings, want %d: %+v", len(findings), len(wantPorts), findings)
	}
	for position, wantPort := range wantPorts {
		if findings[position].Port.Number != wantPort {
			testRun.Errorf("position %d: port %d, want %d", position, findings[position].Port.Number, wantPort)
		}
	}
}
