package domain

import "testing"

const testPortNumber = 80

func TestPortExposure(testRun *testing.T) {
	testCases := []struct {
		name    string
		address string
		want    Exposure
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
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(subRun *testing.T) {
			port := Port{Number: testPortNumber, Protocol: TCP, BindAddress: testCase.address}
			if got := port.Exposure(); got != testCase.want {
				subRun.Fatalf("Exposure(%q) = %v, want %v", testCase.address, got, testCase.want)
			}
		})
	}
}
