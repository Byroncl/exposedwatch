package nmapscan

import (
	"errors"
	"testing"
)

func TestParsePortSpecAccepts(testRun *testing.T) {
	for _, raw := range []string{"22", "22,80,443", "8000-8100", "22,80,443,8000-8100", "1-65535", "65535"} {
		spec, err := ParsePortSpec(raw)
		if err != nil || string(spec) != raw {
			testRun.Errorf("%q: spec=%q err=%v", raw, spec, err)
		}
	}
}

func TestParsePortSpecEmptyMeansTopPorts(testRun *testing.T) {
	spec, err := ParsePortSpec("")
	if err != nil || spec != "" {
		testRun.Fatalf("spec=%q err=%v", spec, err)
	}
}

func TestParsePortSpecRejects(testRun *testing.T) {
	for _, raw := range []string{
		"0", "65536", "-80", "80-", "80--90", "90-80", "22,,80", ",22", "22,", "abc", "22;ls",
		"22 80", "T:22", "U:53", "1-2-3", "$(id)", "22\n", "99999999999999999999", "--script=x",
	} {
		if _, err := ParsePortSpec(raw); !errors.Is(err, ErrInvalidPortSpec) {
			testRun.Errorf("%q should be rejected, got %v", raw, err)
		}
	}
}
