package main

import (
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/tui"
)

func TestTuiRefusesWhenStdoutIsNotATerminal(testRun *testing.T) {
	code, stdout, stderr := runCommand("tui", "--no-docker")
	if code != ExitUsage || stdout != "" || !strings.Contains(stderr, "interactive terminal") {
		testRun.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestTuiRejectsBadFlags(testRun *testing.T) {
	for _, arguments := range [][]string{
		{"tui", "--interval", "500ms"},
		{"tui", "--interval", "0s"},
		{"tui", "--nope"},
		{"tui", "extra"},
	} {
		code, _, stderr := runCommand(arguments...)
		if code != ExitUsage || stderr == "" {
			testRun.Errorf("%v: code=%d stderr=%q", arguments, code, stderr)
		}
	}
}

func TestParseTuiOptionsDefaults(testRun *testing.T) {
	options, err := parseTuiOptions(nil, &strings.Builder{})
	if err != nil {
		testRun.Fatal(err)
	}
	if options.Interval != 2*time.Second || options.ProcRoot != "/proc" || options.NoDocker {
		testRun.Fatalf("options = %+v", options)
	}
}

func TestParseTuiOptionsAcceptsMinimumInterval(testRun *testing.T) {
	options, err := parseTuiOptions([]string{"--interval", "1s", "--no-docker", "--proc-root", "/x"}, &strings.Builder{})
	if err != nil || options.Interval != time.Second || !options.NoDocker || options.ProcRoot != "/x" {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestUsageMentionsTuiCommand(testRun *testing.T) {
	_, _, stderr := runCommand()
	if !strings.Contains(stderr, "tui") {
		testRun.Fatalf("usage lacks tui: %q", stderr)
	}
}

func TestParseTuiOptionsThemeDefaultsToAuto(testRun *testing.T) {
	options, err := parseTuiOptions(nil, &strings.Builder{})
	if err != nil || options.Theme != tui.ThemeAuto {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestParseTuiOptionsAcceptsEveryTheme(testRun *testing.T) {
	cases := map[string]tui.Theme{"auto": tui.ThemeAuto, "dark": tui.ThemeDark, "light": tui.ThemeLight}
	for name, want := range cases {
		options, err := parseTuiOptions([]string{"--theme", name}, &strings.Builder{})
		if err != nil || options.Theme != want {
			testRun.Errorf("--theme %s: options=%+v err=%v", name, options, err)
		}
	}
}

func TestTuiRejectsUnknownTheme(testRun *testing.T) {
	code, _, stderr := runCommand("tui", "--theme", "neon")
	if code != ExitUsage || !strings.Contains(stderr, "--theme") || !strings.Contains(stderr, "neon") {
		testRun.Fatalf("code=%d stderr=%q", code, stderr)
	}
}

func TestUsageMentionsThemeFlag(testRun *testing.T) {
	_, stdout, _ := runCommand("help")
	if !strings.Contains(stdout, "--theme") {
		testRun.Fatalf("usage lacks --theme: %q", stdout)
	}
}

func TestParseTuiOptionsConnectionsAreEnabledByDefault(testRun *testing.T) {
	options, err := parseTuiOptions(nil, &strings.Builder{})
	if err != nil || options.NoConnections {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestParseTuiOptionsAcceptsNoConnections(testRun *testing.T) {
	options, err := parseTuiOptions([]string{"--no-connections"}, &strings.Builder{})
	if err != nil || !options.NoConnections {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestUsageMentionsNoConnectionsFlag(testRun *testing.T) {
	_, stdout, _ := runCommand("help")
	if !strings.Contains(stdout, "--no-connections") {
		testRun.Fatalf("usage lacks --no-connections: %q", stdout)
	}
}

func TestBuildConnectionSourceHonorsTheFlag(testRun *testing.T) {
	if buildConnectionSource(TuiOptions{NoConnections: true}) != nil {
		testRun.Fatal("--no-connections must yield no connection source")
	}
	if buildConnectionSource(TuiOptions{ScanOptions: ScanOptions{ProcRoot: testRun.TempDir()}}) == nil {
		testRun.Fatal("connections must be enabled by default")
	}
}

func TestParseTuiOptionsFirewallIsEnabledByDefaultAndAcceptsNoFirewall(testRun *testing.T) {
	options, err := parseTuiOptions(nil, &strings.Builder{})
	if err != nil || options.NoFirewall {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
	options, err = parseTuiOptions([]string{"--no-firewall"}, &strings.Builder{})
	if err != nil || !options.NoFirewall {
		testRun.Fatalf("options=%+v err=%v", options, err)
	}
}

func TestBuildFirewallEvaluatorHonorsTheFlag(testRun *testing.T) {
	if buildFirewallEvaluator(TuiOptions{NoFirewall: true}) != nil {
		testRun.Fatal("--no-firewall must yield no evaluator, so the firewall is never read")
	}
	if buildFirewallEvaluator(TuiOptions{}) == nil {
		testRun.Fatal("the firewall evaluator must be enabled by default")
	}
}
