package baselinefile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

const validDocument = `{"version":1,"created_at":"2026-11-11T10:00:00Z","tool_version":"v0.4.0","entries":[` +
	`{"protocol":"tcp","port":22,"bind_address":"0.0.0.0","source":"host"}]}` + "\n"

func trustAs(userID int, ownerID int, ownerKnown bool) Trust {
	return Trust{
		EffectiveUserID: userID,
		Lstat:           os.Lstat,
		Stat:            os.Stat,
		OwnerOf:         func(os.FileInfo) (int, bool) { return ownerID, ownerKnown },
	}
}

func writeBaselineFile(testRun *testing.T, directory, content string, mode os.FileMode) ports.BaselinePath {
	testRun.Helper()
	path := filepath.Join(directory, "baseline.json")
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		testRun.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		testRun.Fatal(err)
	}
	return ports.BaselinePath(path)
}

func sampleBaseline() domain.Baseline {
	return domain.Baseline{
		Version:     domain.BaselineVersion,
		CreatedAt:   time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
		ToolVersion: "v0.4.0",
		Entries: []domain.BaselineEntry{
			{Protocol: domain.TCP, Port: 80, BindAddress: "0.0.0.0", Source: domain.SourceHost},
			{Protocol: domain.TCP, Port: 22, BindAddress: "127.0.0.1", Source: domain.SourceHost},
			{Protocol: domain.TCP, Port: 8080, BindAddress: "::", Source: domain.SourceDocker},
		},
	}
}

func ownTrust() Trust { return trustAs(os.Geteuid(), os.Geteuid(), true) }

func filePathIn(directory string) ports.BaselinePath {
	return ports.BaselinePath(filepath.Join(directory, "baseline.json"))
}

// privateTempDir is a temporary directory that does not depend on the umask of the machine.
func privateTempDir(testRun *testing.T) string {
	testRun.Helper()
	directory := testRun.TempDir()
	if err := os.Chmod(directory, 0o700); err != nil {
		testRun.Fatal(err)
	}
	return directory
}
