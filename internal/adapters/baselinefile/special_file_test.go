package baselinefile

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/Byroncl/exposedwatch/internal/ports"
)

func loadWithTimeout(testRun *testing.T, path ports.BaselinePath) error {
	testRun.Helper()
	result := make(chan error, 1)
	go func() {
		_, err := NewStore(ownTrust()).Load(path)
		result <- err
	}()
	select {
	case err := <-result:
		return err
	case <-time.After(3 * time.Second):
		testRun.Fatal("Load hangs on this file type")
		return nil
	}
}

func TestLoadRefusesAFifoWithoutHanging(testRun *testing.T) {
	path := filepath.Join(privateTempDir(testRun), "baseline.json")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		testRun.Fatal(err)
	}
	if err := loadWithTimeout(testRun, ports.BaselinePath(path)); !errors.Is(err, ErrUntrustedBaseline) {
		testRun.Fatalf("got %v", err)
	}
}

func TestLoadRefusesDirectoriesAndDevices(testRun *testing.T) {
	directory := filepath.Join(privateTempDir(testRun), "baseline.json")
	if err := os.Mkdir(directory, 0o700); err != nil {
		testRun.Fatal(err)
	}
	if err := loadWithTimeout(testRun, ports.BaselinePath(directory)); !errors.Is(err, ErrUntrustedBaseline) {
		testRun.Fatalf("directory: got %v", err)
	}
	if err := loadWithTimeout(testRun, ports.BaselinePath(os.DevNull)); !errors.Is(err, ErrUntrustedBaseline) {
		testRun.Fatalf("device: got %v", err)
	}
}

func TestLoadChecksTheParentDirectoryOwner(testRun *testing.T) {
	directory := privateTempDir(testRun)
	path := writeBaselineFile(testRun, directory, validDocument, 0o600)
	ownerFor := func(fileOwner, directoryOwner int) func(os.FileInfo) (int, bool) {
		return func(info os.FileInfo) (int, bool) {
			if info.IsDir() {
				return directoryOwner, true
			}
			return fileOwner, true
		}
	}
	testCases := []struct {
		name         string
		user         int
		directoryUID int
		wantRefused  bool
		wantWarning  bool
	}{
		{"root, directory owned by root", 0, 0, false, false},
		{"root, directory owned by someone else", 0, 1000, true, false},
		{"user, directory owned by the user", 1000, 1000, false, false},
		{"user, directory owned by root", 1000, 0, false, false},
		{"user, directory owned by another user", 1000, 4242, false, true},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			trust := trustAs(testCase.user, testCase.user, true)
			trust.OwnerOf = ownerFor(testCase.user, testCase.directoryUID)
			loaded, err := NewStore(trust).Load(path)
			if testCase.wantRefused != errors.Is(err, ErrUntrustedBaseline) || (!testCase.wantRefused && err != nil) {
				testRun.Fatalf("err=%v", err)
			}
			if !testCase.wantRefused && (len(loaded.Warnings) == 1) != testCase.wantWarning {
				testRun.Fatalf("warnings=%v", loaded.Warnings)
			}
		})
	}
}
