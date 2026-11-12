package configfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func trustAs(userID int, ownerID int, ownerKnown bool) Trust {
	return Trust{
		EffectiveUserID: userID,
		Lstat:           os.Lstat,
		OwnerOf:         func(os.FileInfo) (int, bool) { return ownerID, ownerKnown },
	}
}

func TestLoadRefusesASymlinkedConfig(testRun *testing.T) {
	directory := testRun.TempDir()
	target := writeConfigFile(testRun, directory, `{"version":1,"telegram":{"bot_token":"`+fileToken+`"}}`, 0o600)
	link := filepath.Join(directory, "link.json")
	if err := os.Symlink(target, link); err != nil {
		testRun.Fatal(err)
	}
	_, err := Load(link, noEnvironment)
	if !errors.Is(err, ErrUntrustedConfig) || strings.Contains(err.Error(), "FILE-token") {
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
			path := writeConfigFile(testRun, testRun.TempDir(), `{"version":1}`, 0o600)
			_, err := LoadWithTrust(path, noEnvironment, testCase.trust)
			if testCase.wantErr != errors.Is(err, ErrUntrustedConfig) || (!testCase.wantErr && err != nil) {
				testRun.Fatalf("got %v", err)
			}
		})
	}
}

func TestParentDirectoryWritableByOthers(testRun *testing.T) {
	directory := testRun.TempDir()
	path := writeConfigFile(testRun, directory, `{"version":1}`, 0o600)
	if err := os.Chmod(directory, 0o777); err != nil {
		testRun.Fatal(err)
	}
	if _, err := LoadWithTrust(path, noEnvironment, trustAs(0, 0, true)); !errors.Is(err, ErrUntrustedConfig) {
		testRun.Fatalf("root must refuse, got %v", err)
	}
	loaded, err := LoadWithTrust(path, noEnvironment, trustAs(1000, 1000, true))
	if err != nil || len(loaded.Warnings) != 1 || !strings.Contains(loaded.Warnings[0], "written by group or others") {
		testRun.Fatalf("non-root must only warn: %v %v", err, loaded.Warnings)
	}
	if err := os.Chmod(directory, 0o755); err != nil {
		testRun.Fatal(err)
	}
	if _, err := LoadWithTrust(path, noEnvironment, trustAs(0, 0, true)); err != nil {
		testRun.Fatalf("0755 parent is fine: %v", err)
	}
}

func TestSystemTrustAcceptsAFileOwnedByTheCurrentUser(testRun *testing.T) {
	path := writeConfigFile(testRun, testRun.TempDir(), `{"version":1}`, 0o600)
	if _, err := Load(path, noEnvironment); err != nil {
		testRun.Fatal(err)
	}
}
