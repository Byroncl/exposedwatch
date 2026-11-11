package iptablessave

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const (
	fileModeExecutable    = 0o755
	fileModeNotExecutable = 0o644
)

func writeBinary(testRun *testing.T, directory string, name string, mode os.FileMode) string {
	testRun.Helper()
	if err := os.Chmod(directory, fileModeExecutable); err != nil {
		testRun.Fatal(err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		testRun.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		testRun.Fatal(err)
	}
	return path
}

func failingPathLookup(string) (string, error) {
	return "", errors.New("not in PATH")
}

func TestLocatorFindsBinaryInSbinOnlyDirectory(testRun *testing.T) {
	sbin := testRun.TempDir()
	expected := writeBinary(testRun, sbin, ipv4SaveBinaryName, fileModeExecutable)
	locator := BinaryLocator{Directories: []string{testRun.TempDir(), sbin}, PathLookup: failingPathLookup}
	found, err := locator.Find(ipv4SaveBinaryName)
	if err != nil || found != expected {
		testRun.Fatalf("found=%q err=%v", found, err)
	}
}

func TestLocatorPrefersEarlierDirectory(testRun *testing.T) {
	first := testRun.TempDir()
	second := testRun.TempDir()
	expected := writeBinary(testRun, first, ipv4SaveBinaryName, fileModeExecutable)
	writeBinary(testRun, second, ipv4SaveBinaryName, fileModeExecutable)
	locator := BinaryLocator{Directories: []string{first, second}, PathLookup: failingPathLookup}
	if found, _ := locator.Find(ipv4SaveBinaryName); found != expected {
		testRun.Fatalf("found=%q", found)
	}
}

func TestLocatorIgnoresNonExecutableFilesAndDirectories(testRun *testing.T) {
	directory := testRun.TempDir()
	writeBinary(testRun, directory, ipv4SaveBinaryName, fileModeNotExecutable)
	if err := os.Mkdir(filepath.Join(directory, ipv6SaveBinaryName), fileModeExecutable); err != nil {
		testRun.Fatal(err)
	}
	locator := BinaryLocator{Directories: []string{directory}, PathLookup: failingPathLookup}
	for _, name := range []string{ipv4SaveBinaryName, ipv6SaveBinaryName} {
		if _, err := locator.Find(name); !errors.Is(err, ErrBinaryNotFound) {
			testRun.Errorf("%s: err=%v", name, err)
		}
	}
}

func TestLocatorFallsBackToPathLast(testRun *testing.T) {
	fromPath := writeBinary(testRun, testRun.TempDir(), ipv4SaveBinaryName, fileModeExecutable)
	locator := BinaryLocator{
		Directories: []string{testRun.TempDir()},
		PathLookup:  func(string) (string, error) { return fromPath, nil },
	}
	found, err := locator.Find(ipv4SaveBinaryName)
	if err != nil || found != fromPath {
		testRun.Fatalf("found=%q err=%v", found, err)
	}
}

func TestLocatorMissingEverywhere(testRun *testing.T) {
	locator := BinaryLocator{Directories: []string{testRun.TempDir()}}
	if _, err := locator.Find(ipv4SaveBinaryName); !errors.Is(err, ErrBinaryNotFound) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestDefaultDirectoriesCoverSbin(testRun *testing.T) {
	expected := []string{"/usr/sbin", "/sbin", "/usr/local/sbin", "/usr/bin", "/bin"}
	if len(DefaultBinaryDirectories) != len(expected) {
		testRun.Fatalf("directories = %v", DefaultBinaryDirectories)
	}
	for index, directory := range expected {
		if DefaultBinaryDirectories[index] != directory {
			testRun.Fatalf("directories = %v", DefaultBinaryDirectories)
		}
	}
}

const (
	fileModeGroupWritable = 0o775
	fileModeWorldWritable = 0o777
)

func TestLocatorRejectsWritableDirectory(testRun *testing.T) {
	for _, mode := range []os.FileMode{fileModeGroupWritable, fileModeWorldWritable} {
		directory := testRun.TempDir()
		writeBinary(testRun, directory, ipv4SaveBinaryName, fileModeExecutable)
		if err := os.Chmod(directory, mode); err != nil {
			testRun.Fatal(err)
		}
		locator := BinaryLocator{Directories: []string{directory}, PathLookup: failingPathLookup}
		if _, err := locator.Find(ipv4SaveBinaryName); !errors.Is(err, ErrBinaryNotFound) {
			testRun.Errorf("directory mode %o accepted: %v", mode, err)
		}
	}
}

func TestLocatorRejectsWritableFile(testRun *testing.T) {
	directory := testRun.TempDir()
	path := writeBinary(testRun, directory, ipv4SaveBinaryName, fileModeExecutable)
	if err := os.Chmod(path, fileModeWorldWritable); err != nil {
		testRun.Fatal(err)
	}
	locator := BinaryLocator{Directories: []string{directory}, PathLookup: failingPathLookup}
	if _, err := locator.Find(ipv4SaveBinaryName); !errors.Is(err, ErrBinaryNotFound) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestLocatorSkipsUnsafeDirectoryAndUsesNextSafeOne(testRun *testing.T) {
	unsafe := testRun.TempDir()
	safe := testRun.TempDir()
	writeBinary(testRun, unsafe, ipv4SaveBinaryName, fileModeExecutable)
	expected := writeBinary(testRun, safe, ipv4SaveBinaryName, fileModeExecutable)
	if err := os.Chmod(unsafe, fileModeWorldWritable); err != nil {
		testRun.Fatal(err)
	}
	locator := BinaryLocator{Directories: []string{unsafe, safe}, PathLookup: failingPathLookup}
	if found, err := locator.Find(ipv4SaveBinaryName); err != nil || found != expected {
		testRun.Fatalf("found=%q err=%v", found, err)
	}
}

func TestLocatorAppliesSafetyCheckToPathResult(testRun *testing.T) {
	directory := testRun.TempDir()
	path := writeBinary(testRun, directory, ipv4SaveBinaryName, fileModeExecutable)
	if err := os.Chmod(directory, fileModeWorldWritable); err != nil {
		testRun.Fatal(err)
	}
	locator := BinaryLocator{PathLookup: func(string) (string, error) { return path, nil }}
	if _, err := locator.Find(ipv4SaveBinaryName); !errors.Is(err, ErrBinaryNotFound) {
		testRun.Fatalf("err=%v", err)
	}
}
