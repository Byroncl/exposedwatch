package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/adapters/baselinefile"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

const tcpTableHeader = "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n"

// procListener is one listening TCP socket of a fake /proc/net/tcp table.
type procListener struct {
	hexAddress string
	port       int
}

const (
	wildcardHex = "00000000"
	loopbackHex = "0100007F"
)

// writeProcWithListeners builds a fake proc root holding exactly these sockets.
func writeProcWithListeners(testRun *testing.T, listeners ...procListener) string {
	testRun.Helper()
	procRoot := testRun.TempDir()
	netDirectory := filepath.Join(procRoot, "net")
	if err := os.MkdirAll(netDirectory, 0o755); err != nil {
		testRun.Fatal(err)
	}
	var table strings.Builder
	table.WriteString(tcpTableHeader)
	for index, listener := range listeners {
		fmt.Fprintf(&table, "   %d: %s:%04X 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 %d 1 0000000000000000 100 0 0 10 0\n",
			index, listener.hexAddress, listener.port, 1000+index)
	}
	if err := os.WriteFile(filepath.Join(netDirectory, "tcp"), []byte(table.String()), 0o644); err != nil {
		testRun.Fatal(err)
	}
	return procRoot
}

func storeBaseline(testRun *testing.T, entries ...domain.BaselineEntry) string {
	testRun.Helper()
	directory := testRun.TempDir()
	if err := os.Chmod(directory, 0o700); err != nil {
		testRun.Fatal(err)
	}
	path := filepath.Join(directory, "baseline.json")
	baseline := domain.Baseline{
		Version: domain.BaselineVersion, CreatedAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
		ToolVersion: "test", Entries: entries,
	}
	if err := baselinefile.NewStore(baselinefile.SystemTrust()).Save(ports.BaselinePath(path), baseline); err != nil {
		testRun.Fatal(err)
	}
	return path
}

func hostEntry(port int, bind string) domain.BaselineEntry {
	return domain.BaselineEntry{Protocol: domain.TCP, Port: port, BindAddress: bind, Source: domain.SourceHost}
}

func dockerEntry(port int, bind string) domain.BaselineEntry {
	return domain.BaselineEntry{Protocol: domain.TCP, Port: port, BindAddress: bind, Source: domain.SourceDocker}
}

func executeBaselineWith(arguments []string, dependencies baselineDependencies) (ExitCode, string, string) {
	var stdout, stderr bytes.Buffer
	code := executeBaseline(arguments, &stdout, &stderr, dependencies)
	return code, stdout.String(), stderr.String()
}

func fixedBaselineDependencies(stateHome string) baselineDependencies {
	return baselineDependencies{
		environment: baselinefile.Environment{
			UserID:        1000,
			Getenv:        func(name string) string { return map[string]string{"XDG_STATE_HOME": stateHome}[name] },
			HomeDirectory: func() (string, error) { return "/nonexistent-home", nil },
		},
		now: func() time.Time { return time.Date(2026, 11, 11, 12, 30, 0, 0, time.UTC) },
	}
}

// missingDockerSocket is a socket path that cannot exist, so the Docker source fails without any network.
func missingDockerSocket(testRun *testing.T) string {
	testRun.Helper()
	return filepath.Join(testRun.TempDir(), "no-docker.sock")
}
