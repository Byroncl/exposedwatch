package baselinefile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/ports"
)

func withZeroUmask(testRun *testing.T) {
	testRun.Helper()
	previous := syscall.Umask(0)
	testRun.Cleanup(func() { syscall.Umask(previous) })
}

func TestSaveWritesPrivateFileInPrivateDirectory(testRun *testing.T) {
	withZeroUmask(testRun)
	path := ports.BaselinePath(filepath.Join(privateTempDir(testRun), "nested", "state", "baseline.json"))
	if err := NewStore(ownTrust()).Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	fileInfo, err := os.Stat(string(path))
	if err != nil || fileInfo.Mode().Perm() != 0o600 {
		testRun.Fatalf("file: %v %v", fileInfo, err)
	}
	directoryInfo, err := os.Stat(filepath.Dir(string(path)))
	if err != nil || directoryInfo.Mode().Perm() != 0o700 {
		testRun.Fatalf("directory: %v %v", directoryInfo, err)
	}
	entries, _ := os.ReadDir(filepath.Dir(string(path)))
	if len(entries) != 1 {
		testRun.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestSaveIsDeterministicAndNewlineTerminated(testRun *testing.T) {
	directory := privateTempDir(testRun)
	first := ports.BaselinePath(filepath.Join(directory, "first.json"))
	second := ports.BaselinePath(filepath.Join(directory, "second.json"))
	reversed := sampleBaseline()
	reversed.Entries[0], reversed.Entries[2] = reversed.Entries[2], reversed.Entries[0]
	store := NewStore(ownTrust())
	if err := store.Save(first, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	if err := store.Save(second, reversed); err != nil {
		testRun.Fatal(err)
	}
	firstContent, _ := os.ReadFile(string(first))
	secondContent, _ := os.ReadFile(string(second))
	if string(firstContent) != string(secondContent) || !strings.HasSuffix(string(firstContent), "}\n") {
		testRun.Fatalf("not deterministic or not newline terminated:\n%s\n%s", firstContent, secondContent)
	}
	if !strings.Contains(string(firstContent), `"source": "docker"`) || strings.Index(string(firstContent), `"docker"`) > strings.Index(string(firstContent), `"host"`) {
		testRun.Fatalf("entries must be sorted by source first:\n%s", firstContent)
	}
}

func TestSavedFileHoldsOnlyTheAllowedFields(testRun *testing.T) {
	path := ports.BaselinePath(filepath.Join(privateTempDir(testRun), "baseline.json"))
	if err := NewStore(ownTrust()).Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	content, _ := os.ReadFile(string(path))
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		testRun.Fatal(err)
	}
	allowedTop := map[string]bool{"version": true, "created_at": true, "tool_version": true, "entries": true}
	for key := range document {
		if !allowedTop[key] {
			testRun.Errorf("unexpected top-level key %q", key)
		}
	}
	allowedEntry := map[string]bool{"protocol": true, "port": true, "bind_address": true, "source": true}
	for _, raw := range document["entries"].([]any) {
		for key := range raw.(map[string]any) {
			if !allowedEntry[key] {
				testRun.Errorf("unexpected entry key %q", key)
			}
		}
	}
	for _, forbidden := range []string{"process", "container", "host", "remote", "hostname", "name"} {
		if strings.Contains(string(content), `"`+forbidden+`":`) {
			testRun.Errorf("file must not contain a %q key:\n%s", forbidden, content)
		}
	}
}

func TestSaveEmptyBaselineWritesAnEmptyArray(testRun *testing.T) {
	path := ports.BaselinePath(filepath.Join(privateTempDir(testRun), "baseline.json"))
	empty := sampleBaseline()
	empty.Entries = nil
	if err := NewStore(ownTrust()).Save(path, empty); err != nil {
		testRun.Fatal(err)
	}
	content, _ := os.ReadFile(string(path))
	if !strings.Contains(string(content), `"entries": []`) {
		testRun.Fatalf("got:\n%s", content)
	}
}

func TestSaveRejectsAnInvalidBaselineBeforeTouchingDisk(testRun *testing.T) {
	path := ports.BaselinePath(filepath.Join(privateTempDir(testRun), "baseline.json"))
	invalid := sampleBaseline()
	invalid.Entries[0].Port = 0
	if err := NewStore(ownTrust()).Save(path, invalid); err == nil {
		testRun.Fatal("expected an error")
	}
	if _, err := os.Stat(string(path)); !errors.Is(err, os.ErrNotExist) {
		testRun.Fatalf("nothing may be written, stat=%v", err)
	}
}

func TestFailedRenameKeepsThePreviousFileAndCleansUp(testRun *testing.T) {
	directory := privateTempDir(testRun)
	path := writeBaselineFile(testRun, directory, validDocument, 0o600)
	failingRename := func(string, string) error { return errors.New("rename refused") }
	err := writeAtomically(string(path), []byte("new content"), failingRename)
	if !errors.Is(err, ErrWriteBaseline) {
		testRun.Fatalf("got %v", err)
	}
	content, _ := os.ReadFile(string(path))
	if string(content) != validDocument {
		testRun.Fatalf("previous file was damaged: %q", content)
	}
	entries, _ := os.ReadDir(directory)
	if len(entries) != 1 {
		testRun.Fatalf("temporary file left behind: %v", entries)
	}
}

func TestSaveRepairsLoosePermissionsOfAnOlderFile(testRun *testing.T) {
	directory := privateTempDir(testRun)
	path := writeBaselineFile(testRun, directory, validDocument, 0o644)
	if err := NewStore(ownTrust()).Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	info, _ := os.Stat(string(path))
	if info.Mode().Perm() != 0o600 {
		testRun.Fatalf("mode %v", info.Mode().Perm())
	}
}
