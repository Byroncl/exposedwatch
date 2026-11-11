package nmapscan

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const testdataDirectory = "testdata"

func readFixture(testRun *testing.T, name string) []byte {
	testRun.Helper()
	content, err := os.ReadFile(filepath.Join(testdataDirectory, name))
	if err != nil {
		testRun.Fatal(err)
	}
	return content
}

func TestParseResultSingleHostKeepsOnlyOpenPorts(testRun *testing.T) {
	openPorts, err := parseResult(readFixture(testRun, "open_ports_single_host.xml"))
	if err != nil {
		testRun.Fatal(err)
	}
	want := []openPort{{Address: "192.168.1.10", Number: 22}, {Address: "192.168.1.10", Number: 5432}}
	if len(openPorts) != len(want) {
		testRun.Fatalf("got %v", openPorts)
	}
	for index := range want {
		if openPorts[index] != want[index] {
			testRun.Errorf("index %d: got %v want %v", index, openPorts[index], want[index])
		}
	}
}

func TestParseResultMultipleHostsSkipsMacUdpAndClosed(testRun *testing.T) {
	openPorts, err := parseResult(readFixture(testRun, "multiple_hosts.xml"))
	if err != nil {
		testRun.Fatal(err)
	}
	want := []openPort{{Address: "10.0.0.5", Number: 80}, {Address: "2001:db8::1", Number: 443}}
	if len(openPorts) != len(want) {
		testRun.Fatalf("got %v", openPorts)
	}
	for index := range want {
		if openPorts[index] != want[index] {
			testRun.Errorf("index %d: got %v want %v", index, openPorts[index], want[index])
		}
	}
}

func TestParseResultNoHostsIsEmptyNotError(testRun *testing.T) {
	openPorts, err := parseResult(readFixture(testRun, "no_hosts.xml"))
	if err != nil || len(openPorts) != 0 {
		testRun.Fatalf("ports=%v err=%v", openPorts, err)
	}
}

func TestParseResultRejectsBadOutput(testRun *testing.T) {
	for _, name := range []string{"malformed.xml", "wrong_root.xml"} {
		if _, err := parseResult(readFixture(testRun, name)); !errors.Is(err, ErrInvalidOutput) {
			testRun.Errorf("%s: err=%v", name, err)
		}
	}
	if _, err := parseResult(nil); !errors.Is(err, ErrInvalidOutput) {
		testRun.Errorf("empty: err=%v", err)
	}
}
