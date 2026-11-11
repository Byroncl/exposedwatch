package baselinefile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

func TestLoadRoundTrip(testRun *testing.T) {
	path := filePathIn(privateTempDir(testRun))
	store := NewStore(ownTrust())
	if err := store.Save(path, sampleBaseline()); err != nil {
		testRun.Fatal(err)
	}
	loaded, err := store.Load(path)
	if err != nil {
		testRun.Fatal(err)
	}
	want := sampleBaseline().Normalized()
	if loaded.Baseline.Version != want.Version || !loaded.Baseline.CreatedAt.Equal(want.CreatedAt) || loaded.Baseline.ToolVersion != want.ToolVersion {
		testRun.Fatalf("header: %+v", loaded.Baseline)
	}
	if fmt.Sprint(loaded.Baseline.Entries) != fmt.Sprint(want.Entries) {
		testRun.Fatalf("entries: %v want %v", loaded.Baseline.Entries, want.Entries)
	}
	if len(loaded.Warnings) != 0 {
		testRun.Fatalf("warnings: %v", loaded.Warnings)
	}
}

func TestLoadMissingFileIsReported(testRun *testing.T) {
	_, err := NewStore(ownTrust()).Load(filePathIn(privateTempDir(testRun)))
	if !errors.Is(err, ErrMissingBaseline) {
		testRun.Fatalf("got %v", err)
	}
}

func TestLoadCollapsesDuplicates(testRun *testing.T) {
	entry := `{"protocol":"tcp","port":22,"bind_address":"0.0.0.0","source":"host"}`
	document := `{"version":1,"created_at":"2026-11-11T10:00:00Z","tool_version":"v","entries":[` + entry + "," + entry + `]}`
	path := writeBaselineFile(testRun, privateTempDir(testRun), document, 0o600)
	loaded, err := NewStore(ownTrust()).Load(path)
	if err != nil || len(loaded.Baseline.Entries) != 1 {
		testRun.Fatalf("err=%v entries=%v", err, loaded.Baseline.Entries)
	}
}

func TestLoadRejectsBadContentWithoutEchoingIt(testRun *testing.T) {
	const leak = "LEAK-MARKER"
	header := `"version":1,"created_at":"2026-11-11T10:00:00Z","tool_version":"v"`
	entry := func(fields string) string { return `{` + header + `,"entries":[{` + fields + `}]}` }
	hugeEntries := strings.Repeat(`{"protocol":"tcp","port":1,"bind_address":"::","source":"host"},`, domain.MaxBaselineEntries) +
		`{"protocol":"tcp","port":1,"bind_address":"::","source":"host"}`
	testCases := []struct {
		name    string
		content string
		wantErr error
	}{
		{"not json", leak + " not json", ErrMalformedBaseline},
		{"empty file", "", ErrMalformedBaseline},
		{"trailing data", validDocument + `{"x":"` + leak + `"}`, ErrMalformedBaseline},
		{"unknown top-level field", `{` + header + `,"entries":[],"` + leak + `":1}`, ErrMalformedBaseline},
		{"unknown entry field", entry(`"protocol":"tcp","port":22,"bind_address":"0.0.0.0","source":"host","process":"` + leak + `"`), ErrMalformedBaseline},
		{"wrong type", `{"version":"` + leak + `"}`, ErrMalformedBaseline},
		{"wrong version", `{"version":2,"created_at":"2026-11-11T10:00:00Z","tool_version":"v","entries":[]}`, domain.ErrInvalidBaseline},
		{"missing version", `{"created_at":"2026-11-11T10:00:00Z","tool_version":"v","entries":[]}`, domain.ErrInvalidBaseline},
		{"bad timestamp", `{"version":1,"created_at":"` + leak + `","tool_version":"v","entries":[]}`, domain.ErrInvalidBaseline},
		{"bad protocol", entry(`"protocol":"` + leak + `","port":22,"bind_address":"0.0.0.0","source":"host"`), domain.ErrInvalidBaseline},
		{"bad port", entry(`"protocol":"tcp","port":0,"bind_address":"0.0.0.0","source":"host"`), domain.ErrInvalidBaseline},
		{"bad bind", entry(`"protocol":"tcp","port":22,"bind_address":"` + leak + `","source":"host"`), domain.ErrInvalidBaseline},
		{"bad source", entry(`"protocol":"tcp","port":22,"bind_address":"0.0.0.0","source":"external"`), domain.ErrInvalidBaseline},
		{"too many entries", `{` + header + `,"entries":[` + hugeEntries + `]}`, domain.ErrInvalidBaseline},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			path := writeBaselineFile(testRun, privateTempDir(testRun), testCase.content, 0o600)
			_, err := NewStore(ownTrust()).Load(path)
			if !errors.Is(err, testCase.wantErr) {
				testRun.Fatalf("got %v want %v", err, testCase.wantErr)
			}
			if strings.Contains(err.Error(), leak) {
				testRun.Fatalf("error echoes file content: %v", err)
			}
		})
	}
}

func TestLoadRejectsAnOversizedFile(testRun *testing.T) {
	content := validDocument + strings.Repeat(" ", MaxFileBytes)
	path := writeBaselineFile(testRun, privateTempDir(testRun), content, 0o600)
	_, err := NewStore(ownTrust()).Load(path)
	if !errors.Is(err, ErrMalformedBaseline) || !strings.Contains(err.Error(), "larger") {
		testRun.Fatalf("got %v", err)
	}
}

func TestLoadRefusesASymlink(testRun *testing.T) {
	directory := privateTempDir(testRun)
	target := writeBaselineFile(testRun, directory, validDocument, 0o600)
	link := filepath.Join(directory, "link.json")
	if err := os.Symlink(string(target), link); err != nil {
		testRun.Fatal(err)
	}
	_, err := NewStore(ownTrust()).Load(ports.BaselinePath(link))
	if !errors.Is(err, ErrUntrustedBaseline) {
		testRun.Fatalf("got %v", err)
	}
}

func TestLoadRefusesAFileSwappedBetweenLstatAndOpen(testRun *testing.T) {
	directory := privateTempDir(testRun)
	path := writeBaselineFile(testRun, directory, validDocument, 0o600)
	other := writeBaselineFile(testRun, privateTempDir(testRun), validDocument, 0o600)
	otherInfo, err := os.Lstat(string(other))
	if err != nil {
		testRun.Fatal(err)
	}
	trust := ownTrust()
	trust.Lstat = func(string) (os.FileInfo, error) { return otherInfo, nil }
	_, err = NewStore(trust).Load(path)
	if !errors.Is(err, ErrUntrustedBaseline) {
		testRun.Fatalf("got %v", err)
	}
}

func TestLoadChecksTheOwner(testRun *testing.T) {
	testCases := []struct {
		name    string
		trust   Trust
		wantErr bool
	}{
		{"owned by the effective user", trustAs(1000, 1000, true), false},
		{"owned by someone else", trustAs(1000, 4242, true), true},
		{"root run, file owned by another user", trustAs(0, 1000, true), true},
		{"root run, file owned by root", trustAs(0, 0, true), false},
		{"owner cannot be determined", trustAs(1000, 0, false), true},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			path := writeBaselineFile(testRun, privateTempDir(testRun), validDocument, 0o600)
			_, err := NewStore(testCase.trust).Load(path)
			if testCase.wantErr != errors.Is(err, ErrUntrustedBaseline) || (!testCase.wantErr && err != nil) {
				testRun.Fatalf("got %v", err)
			}
		})
	}
}

func TestLoadRefusesAFileOthersCanWrite(testRun *testing.T) {
	for _, mode := range []os.FileMode{0o620, 0o602, 0o666} {
		path := writeBaselineFile(testRun, privateTempDir(testRun), validDocument, mode)
		if _, err := NewStore(ownTrust()).Load(path); !errors.Is(err, ErrUntrustedBaseline) {
			testRun.Errorf("mode %04o: got %v", mode, err)
		}
	}
	path := writeBaselineFile(testRun, privateTempDir(testRun), validDocument, 0o644)
	if _, err := NewStore(ownTrust()).Load(path); err != nil {
		testRun.Errorf("a world-readable file is fine: %v", err)
	}
}

func TestParentDirectoryWritableByOthers(testRun *testing.T) {
	directory := privateTempDir(testRun)
	path := writeBaselineFile(testRun, directory, validDocument, 0o600)
	for _, mode := range []os.FileMode{0o770, 0o707, 0o777} {
		if err := os.Chmod(directory, mode); err != nil {
			testRun.Fatal(err)
		}
		_, err := NewStore(trustAs(0, 0, true)).Load(path)
		if !errors.Is(err, ErrUntrustedBaseline) {
			testRun.Errorf("root, directory %04o: got %v", mode, err)
		}
		loaded, err := NewStore(trustAs(1000, 1000, true)).Load(path)
		if err != nil || len(loaded.Warnings) != 1 || !strings.Contains(loaded.Warnings[0], directory) {
			testRun.Errorf("non-root, directory %04o: err=%v warnings=%v", mode, err, loaded.Warnings)
		}
	}
	if err := os.Chmod(directory, 0o755); err != nil {
		testRun.Fatal(err)
	}
	if loaded, err := NewStore(trustAs(0, 0, true)).Load(path); err != nil || len(loaded.Warnings) != 0 {
		testRun.Fatalf("a 0755 directory is fine: err=%v warnings=%v", err, loaded.Warnings)
	}
}
