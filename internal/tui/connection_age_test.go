package tui

import (
	"testing"
	"time"
)

func TestFormatConnectionAge(testRun *testing.T) {
	cases := map[time.Duration]string{
		-time.Second:                   "0s",
		0:                              "0s",
		12 * time.Second:               "12s",
		59 * time.Second:               "59s",
		time.Minute:                    "1m",
		4*time.Minute + 59*time.Second: "4m",
		time.Hour:                      "1h",
		23*time.Hour + 59*time.Minute:  "23h",
		24 * time.Hour:                 "1d",
		50 * time.Hour:                 "2d",
	}
	for age, want := range cases {
		if got := formatConnectionAge(age); got != want {
			testRun.Errorf("formatConnectionAge(%v) = %q, want %q", age, got, want)
		}
	}
}

func TestConnectionSortModeCycles(testRun *testing.T) {
	mode := ConnectionSortByDirection
	want := []ConnectionSortMode{
		ConnectionSortByLocalPort, ConnectionSortByRemoteAddress, ConnectionSortByProcess, ConnectionSortByAge, ConnectionSortByDirection,
	}
	for _, expected := range want {
		mode = mode.Next()
		if mode != expected {
			testRun.Fatalf("mode = %v, want %v", mode, expected)
		}
	}
	names := map[ConnectionSortMode]string{
		ConnectionSortByDirection: "direction", ConnectionSortByLocalPort: "local-port",
		ConnectionSortByRemoteAddress: "remote", ConnectionSortByProcess: "process", ConnectionSortByAge: "age",
	}
	for sortMode, name := range names {
		if sortMode.String() != name {
			testRun.Errorf("%d name = %q, want %q", sortMode, sortMode.String(), name)
		}
	}
}
