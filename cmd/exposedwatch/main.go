// Command exposedwatch shows what a server exposes to the network.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

// isTopLevelFlag reports whether the argument is handled by the root command
// itself instead of being passed on to the tui.
func isTopLevelFlag(argument string) bool {
	switch argument {
	case flagHelpShort, flagHelpLong, flagVersionLong:
		return true
	default:
		return false
	}
}

// run dispatches the subcommand and returns the process exit code.
func run(arguments []string, stdout, stderr io.Writer) ExitCode {
	// No command, or only flags (for example "--interval 5s"): open the tui.
	if len(arguments) == 0 || (strings.HasPrefix(arguments[0], flagPrefix) && !isTopLevelFlag(arguments[0])) {
		return runTui(arguments, stdout, stderr)
	}
	switch arguments[0] {
	case commandHelp, flagHelpShort, flagHelpLong:
		fmt.Fprint(stdout, usageText)
		return ExitOK
	case flagVersionLong:
		fmt.Fprintf(stdout, versionFormat, version)
		return ExitOK
	case commandScan:
		return runScan(arguments[1:], stdout, stderr)
	case commandWatch:
		return runWatch(arguments[1:], stdout, stderr)
	case commandNotifyTest:
		return runNotifyTest(arguments[1:], stdout, stderr)
	case commandTui:
		return runTui(arguments[1:], stdout, stderr)
	case commandVersion:
		fmt.Fprintf(stdout, versionFormat, version)
		return ExitOK
	default:
		fmt.Fprint(stderr, usageText)
		return ExitUsage
	}
}
