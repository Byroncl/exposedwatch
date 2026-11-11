package domain

import "testing"

func TestPortExposure(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want Exposure
	}{
		{"ipv4 loopback", "127.0.0.1", Loopback},
		{"ipv4 loopback range", "127.5.5.5", Loopback},
		{"ipv6 loopback", "::1", Loopback},
		{"ipv4 wildcard", "0.0.0.0", AllInterfaces},
		{"ipv6 wildcard", "::", AllInterfaces},
		{"empty", "", AllInterfaces},
		{"star", "*", AllInterfaces},
		{"rfc1918 10/8", "10.1.2.3", Private},
		{"rfc1918 172.16/12", "172.20.0.5", Private},
		{"rfc1918 192.168/16", "192.168.1.10", Private},
		{"ipv6 ula", "fd12:3456::1", Private},
		{"ipv6 with zone", "fe80::1%eth0", Private},
		{"public ipv4", "8.8.8.8", Public},
		{"public ipv6", "2001:4860:4860::8888", Public},
		{"invalid is treated as widest", "not-an-ip", AllInterfaces},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Port{Number: 80, Protocol: TCP, BindAddress: tt.addr}
			if got := p.Exposure(); got != tt.want {
				t.Fatalf("Exposure(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

func TestSeverityOrderingAndString(t *testing.T) {
	order := []Severity{Info, Low, Medium, High, Critical}
	names := []string{"info", "low", "medium", "high", "critical"}
	for i, s := range order {
		if s.String() != names[i] {
			t.Errorf("String() = %q, want %q", s.String(), names[i])
		}
		if i > 0 && order[i-1] >= s {
			t.Errorf("%v should be lower than %v", order[i-1], s)
		}
	}
	if Severity(99).String() != "unknown" {
		t.Errorf("unexpected string for invalid severity")
	}
}

func TestExposureString(t *testing.T) {
	want := map[Exposure]string{Loopback: "loopback", Private: "private", AllInterfaces: "all-interfaces", Public: "public"}
	for e, s := range want {
		if e.String() != s {
			t.Errorf("%d.String() = %q, want %q", e, e.String(), s)
		}
	}
}
