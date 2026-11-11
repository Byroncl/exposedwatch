package nmapscan

import (
	"errors"
	"strings"
	"testing"
)

func TestParseTargetsAcceptsSafeTargets(testRun *testing.T) {
	accepted := []string{
		"192.168.1.10", "10.0.0.0/24", "10.0.0.0/30",
		"example.com", "host.local", "server1", "example.be", "x.ae", "foo.cc", "my-host.internal.example.org", "localhost", "HOST.Example.COM", "127.0.0.1",
	}
	for _, rawTarget := range accepted {
		targets, err := ParseTargets([]string{rawTarget})
		if err != nil {
			testRun.Errorf("%q rejected: %v", rawTarget, err)
			continue
		}
		if len(targets) != 1 {
			testRun.Errorf("%q: got %v", rawTarget, targets)
		}
	}
}

func TestParseTargetsLowercasesHostnames(testRun *testing.T) {
	targets, err := ParseTargets([]string{"HOST.Example.COM"})
	if err != nil || targets[0] != "host.example.com" {
		testRun.Fatalf("targets=%v err=%v", targets, err)
	}
}

func TestParseTargetsRejectsMaliciousOrUnsafeTargets(testRun *testing.T) {
	rejected := map[string]string{
		"option injection":       "-oN=/etc/x",
		"script option":          "--script=x",
		"single dash":            "-",
		"shell separator":        "1.1.1.1; rm -rf /",
		"command substitution":   "$(id)",
		"backticks":              "`id`",
		"newline":                "1.1.1.1\n-oN=/tmp/x",
		"carriage return":        "1.1.1.1\r",
		"tab":                    "1.1.1.1\t2.2.2.2",
		"space":                  "1.1.1.1 2.2.2.2",
		"empty":                  "",
		"blank":                  "   ",
		"quote":                  "'1.1.1.1'",
		"double quote":           "\"example.com\"",
		"at sign":                "user@example.com",
		"colon host":             "example.com:80",
		"pipe":                   "example.com|cat",
		"ampersand":              "example.com&",
		"redirect":               "example.com>/tmp/x",
		"leading dot":            ".example.com",
		"trailing dot":           "example.com.",
		"double dot":             "example..com",
		"leading hyphen label":   "example.-bad.com",
		"trailing hyphen label":  "bad-.example.com",
		"nmap octet range":       "192.168.0-255.1",
		"nmap octet list":        "192.168.0,1.1",
		"out of range ip":        "256.1.1.1",
		"numeric hostname":       "1.2.3",
		"ipv6 zone":              "fe80::1%eth0",
		"unicode hostname":       "exämple.com",
		"slash path":             "example.com/etc",
		"huge ipv4 cidr":         "10.0.0.0/8",
		"wide ipv4 cidr":         "10.0.0.0/23",
		"zero cidr":              "0.0.0.0/0",
		"ipv6 loopback":          "::1",
		"ipv6 global":            "2001:db8::1",
		"ipv4 mapped ipv6":       "::ffff:127.0.0.1",
		"ipv6 cidr":              "2001:db8::/120",
		"huge ipv6 cidr":         "2001:db8::/32",
		"hex literal":            "0x7f000001",
		"hex dotted":             "1.0x1",
		"hex octets":             "0x7f.0.0.1",
		"short ipv4":             "127.1",
		"decimal ipv4":           "2130706433",
		"digit starting tld":     "a.0x1",
		"bad prefix":             "10.0.0.0/abc",
		"too long hostname":      strings.Repeat("a", 64) + ".example.com",
		"too long total":         strings.Repeat("a.", 130) + "com",
		"nul byte":               "example.com\x00",
		"hostname with asterisk": "*.example.com",
	}
	for name, rawTarget := range rejected {
		_, err := ParseTargets([]string{rawTarget})
		if !errors.Is(err, ErrInvalidTarget) {
			testRun.Errorf("%s: %q should be ErrInvalidTarget, got %v", name, rawTarget, err)
		}
	}
}

func TestParseTargetsRejectsEmptyListAndOversizedList(testRun *testing.T) {
	if _, err := ParseTargets(nil); !errors.Is(err, ErrNoTargets) {
		testRun.Fatalf("empty list: %v", err)
	}
	many := make([]string, 0, maxTargetCount+1)
	for index := 0; index <= maxTargetCount; index++ {
		many = append(many, "example.com")
	}
	if _, err := ParseTargets(many); !errors.Is(err, ErrInvalidTarget) {
		testRun.Fatalf("oversized list: %v", err)
	}
}

func TestParseTargetsRejectsListWithOneBadEntry(testRun *testing.T) {
	_, err := ParseTargets([]string{"10.0.0.1", "--script=x"})
	if !errors.Is(err, ErrInvalidTarget) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestTargetListValidateRechecksEveryEntry(testRun *testing.T) {
	if err := (TargetList{"10.0.0.1", "-oN=x"}).Validate(); !errors.Is(err, ErrInvalidTarget) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestParseTargetsReportsIPv6AsUnsupported(testRun *testing.T) {
	for _, rawTarget := range []string{"::1", "2001:db8::/120", "::ffff:127.0.0.1"} {
		_, err := ParseTargets([]string{rawTarget})
		if !errors.Is(err, ErrInvalidTarget) || !strings.Contains(err.Error(), reasonIPv6NotSupported) {
			testRun.Errorf("%q: err=%v", rawTarget, err)
		}
	}
}
