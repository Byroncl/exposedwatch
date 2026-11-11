// Command exposedwatch shows what a server exposes to the network.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(int(run(os.Args[1:], os.Stdout, os.Stderr)))
}

// run dispatches the subcommand and returns the process exit code.
func run(arguments []string, stdout, stderr io.Writer) ExitCode {
	if len(arguments) == 0 {
		fmt.Fprint(stderr, usageText)
		return ExitUsage
	}
	switch arguments[0] {
	case commandScan:
		return runScan(arguments[1:], stdout, stderr)
	case commandVersion:
		fmt.Fprintf(stdout, versionFormat, version)
		return ExitOK
	default:
		fmt.Fprint(stderr, usageText)
		return ExitUsage
	}
}
