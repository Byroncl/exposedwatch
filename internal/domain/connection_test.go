package domain

import "testing"

func TestConnectionDirectionString(testRun *testing.T) {
	testCases := map[ConnectionDirection]string{
		ConnectionInbound:       "inbound",
		ConnectionOutbound:      "outbound",
		ConnectionDirection(99): "unknown",
	}
	for direction, expected := range testCases {
		if direction.String() != expected {
			testRun.Errorf("direction %d = %q, want %q", int(direction), direction.String(), expected)
		}
	}
}

func TestConnectionStateString(testRun *testing.T) {
	testCases := map[ConnectionState]string{
		ConnectionEstablished: "established",
		ConnectionState(99):   "unknown",
	}
	for state, expected := range testCases {
		if state.String() != expected {
			testRun.Errorf("state %d = %q, want %q", int(state), state.String(), expected)
		}
	}
}

func TestConnectionIsLoopbackToLoopback(testRun *testing.T) {
	testCases := []struct {
		name          string
		localAddress  string
		remoteAddress string
		expected      bool
	}{
		{"both ipv4 loopback", "127.0.0.1", "127.0.0.1", true},
		{"both ipv6 loopback", "::1", "::1", true},
		{"mapped loopback", "::ffff:127.0.0.1", "127.0.0.1", true},
		{"remote public", "127.0.0.1", "203.0.113.10", false},
		{"local private remote loopback", "10.0.0.5", "127.0.0.1", false},
		{"unparseable local", "garbage", "127.0.0.1", false},
		{"unparseable remote", "127.0.0.1", "", false},
	}
	for _, testCase := range testCases {
		connection := Connection{LocalAddress: testCase.localAddress, RemoteAddress: testCase.remoteAddress}
		if connection.IsLoopbackToLoopback() != testCase.expected {
			testRun.Errorf("%s: got %v, want %v", testCase.name, !testCase.expected, testCase.expected)
		}
	}
}

func TestNormalizeAddress(testRun *testing.T) {
	testCases := map[string]string{
		"::ffff:203.0.113.10": "203.0.113.10",
		"203.0.113.10":        "203.0.113.10",
		"2001:db8::1":         "2001:db8::1",
		"not-an-address":      "not-an-address",
	}
	for input, expected := range testCases {
		if normalized := NormalizeAddress(input); normalized != expected {
			testRun.Errorf("NormalizeAddress(%q) = %q, want %q", input, normalized, expected)
		}
	}
}
