package baselinefile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/ports"
)

const unrelatedContent = "root:x:0:0:root:/root:/bin/bash\n"

func TestSaveRefusesToOverwriteAFileThatIsNotABaseline(testRun *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{"unrelated text file", unrelatedContent},
		{"empty file", ""},
		{"corrupt baseline", `{"version":1,"entries":`},
		{"baseline with an invalid entry", `{"version":1,"created_at":"2026-11-11T10:00:00Z","tool_version":"v","entries":[{"protocol":"x","port":1,"bind_address":"::","source":"host"}]}`},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			path := writeBaselineFile(testRun, privateTempDir(testRun), testCase.content, 0o644)
			err := NewStore(ownTrust()).Save(path, sampleBaseline())
			if !errors.Is(err, ErrNotABaseline) {
				testRun.Fatalf("got %v", err)
			}
			if strings.Contains(err.Error(), "root:x") {
				testRun.Fatalf("error echoes contents: %v", err)
			}
			content, _ := os.ReadFile(string(path))
			if string(content) != testCase.content {
				testRun.Fatalf("file was modified: %q", content)
			}
		})
	}
}

func TestSaveForceReplacesACorruptFileButNotAnythingElse(testRun *testing.T) {
	path := writeBaselineFile(testRun, privateTempDir(testRun), unrelatedContent, 0o644)
	if err := NewStore(ownTrust()).Forcing().Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	loaded, err := NewStore(ownTrust()).Load(path)
	if err != nil || len(loaded.Baseline.Entries) != 3 {
		testRun.Fatalf("err=%v entries=%v", err, loaded.Baseline.Entries)
	}
}

func TestSaveReplacesAnExistingValidBaseline(testRun *testing.T) {
	path := writeBaselineFile(testRun, privateTempDir(testRun), validDocument, 0o600)
	if err := NewStore(ownTrust()).Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	loaded, err := NewStore(ownTrust()).Load(path)
	if err != nil || len(loaded.Baseline.Entries) != 3 {
		testRun.Fatalf("err=%v", err)
	}
}

func TestSaveRefusesAnUntrustedExistingBaselineUnlessForced(testRun *testing.T) {
	path := writeBaselineFile(testRun, privateTempDir(testRun), validDocument, 0o600)
	foreign := trustAs(os.Geteuid(), os.Geteuid()+1, true)
	if err := NewStore(foreign).Save(path, sampleBaseline()); !errors.Is(err, ErrNotABaseline) {
		testRun.Fatalf("got %v", err)
	}
	if err := NewStore(foreign).Forcing().Save(path, sampleBaseline()); err != nil {
		testRun.Fatalf("force bypasses the content check: %v", err)
	}
}

func TestSaveNeverReplacesASymlinkNotEvenForced(testRun *testing.T) {
	directory := privateTempDir(testRun)
	target := writeBaselineFile(testRun, directory, validDocument, 0o600)
	link := filepath.Join(directory, "link.json")
	if err := os.Symlink(string(target), link); err != nil {
		testRun.Fatal(err)
	}
	for _, store := range []Store{NewStore(ownTrust()), NewStore(ownTrust()).Forcing()} {
		if err := store.Save(ports.BaselinePath(link), sampleBaseline()); !errors.Is(err, ErrNotABaseline) {
			testRun.Fatalf("got %v", err)
		}
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		testRun.Fatal("the symlink must stay in place")
	}
	if content, _ := os.ReadFile(string(target)); string(content) != validDocument {
		testRun.Fatal("the symlink target must not be written")
	}
}

func TestSaveRefusesADirectoryTargetNotEvenForced(testRun *testing.T) {
	directory := filepath.Join(privateTempDir(testRun), "baseline.json")
	if err := os.Mkdir(directory, 0o700); err != nil {
		testRun.Fatal(err)
	}
	for _, store := range []Store{NewStore(ownTrust()), NewStore(ownTrust()).Forcing()} {
		if err := store.Save(ports.BaselinePath(directory), sampleBaseline()); !errors.Is(err, ErrNotABaseline) {
			testRun.Fatalf("got %v", err)
		}
	}
}

func TestSaveCreatesAMissingTarget(testRun *testing.T) {
	path := filePathIn(privateTempDir(testRun))
	if err := NewStore(ownTrust()).Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
}
