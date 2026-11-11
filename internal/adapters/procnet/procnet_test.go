package procnet

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

var _ ports.PortSource = (*Reader)(nil)

const (
	fixtureDirectory   = "testdata"
	fixtureFilePerm    = 0o644
	fixtureDirPerm     = 0o755
	sshdProcessID      = "4242"
	sshdSocketInode    = "12345"
	sshdProcessName    = "sshd"
	sshdPortNumber     = 22
	unreadableProcess  = "999"
	unrelatedLinkPath  = "/dev/null"
	fullFixturePorts   = 6
	withoutUdp6Ports   = 5
	socketLinkTemplate = socketLinkPrefix + sshdSocketInode + socketLinkSuffix
)

var fixtureTables = []ProcFileName{tcpTableName, tcp6TableName, udpTableName, udp6TableName}

type portKey struct {
	protocol    domain.Protocol
	bindAddress string
	number      int
}

func copyFixtures(testRun *testing.T) string {
	testRun.Helper()
	procRoot := testRun.TempDir()
	if err := os.MkdirAll(filepath.Join(procRoot, string(netDirectoryName)), fixtureDirPerm); err != nil {
		testRun.Fatal(err)
	}
	for _, tableName := range fixtureTables {
		relativePath := filepath.Join(string(netDirectoryName), string(tableName))
		content, err := os.ReadFile(filepath.Join(fixtureDirectory, relativePath))
		if err != nil {
			testRun.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(procRoot, relativePath), content, fixtureFilePerm); err != nil {
			testRun.Fatal(err)
		}
	}
	return procRoot
}

func TestListPorts(testRun *testing.T) {
	procRoot := copyFixtures(testRun)
	listedPorts, err := New(procRoot).ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	expectedPorts := map[portKey]bool{
		{domain.TCP, "0.0.0.0", sshdPortNumber}: true,
		{domain.TCP, "127.0.0.1", 631}:          true,
		{domain.TCP, "::", 8080}:                true,
		{domain.TCP, "::1", 3306}:               true,
		{domain.UDP, "0.0.0.0", 53}:             true,
		{domain.UDP, "::", 5353}:                true,
	}
	if len(listedPorts) != len(expectedPorts) {
		testRun.Fatalf("got %d ports, want %d: %+v", len(listedPorts), len(expectedPorts), listedPorts)
	}
	for _, listedPort := range listedPorts {
		key := portKey{listedPort.Protocol, listedPort.BindAddress, listedPort.Number}
		if !expectedPorts[key] {
			testRun.Errorf("unexpected port %+v", listedPort)
		}
		if listedPort.Source != domain.SourceHost {
			testRun.Errorf("source = %q, want host", listedPort.Source)
		}
		if listedPort.Process != "" {
			testRun.Errorf("process = %q, want empty without process directories", listedPort.Process)
		}
	}
}

func TestListPortsMissingTableTolerated(testRun *testing.T) {
	procRoot := copyFixtures(testRun)
	missingTable := filepath.Join(procRoot, string(netDirectoryName), string(udp6TableName))
	if err := os.Remove(missingTable); err != nil {
		testRun.Fatal(err)
	}
	listedPorts, err := New(procRoot).ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if len(listedPorts) != withoutUdp6Ports {
		testRun.Fatalf("got %d ports, want %d", len(listedPorts), withoutUdp6Ports)
	}
}

func TestListPortsNoTableIsError(testRun *testing.T) {
	if _, err := New(testRun.TempDir()).ListPorts(context.Background()); err == nil {
		testRun.Fatal("expected error when no table is readable")
	}
}

func TestListPortsProcessName(testRun *testing.T) {
	procRoot := copyFixtures(testRun)
	descriptorDirectory := filepath.Join(procRoot, sshdProcessID, string(fdDirectoryName))
	if err := os.MkdirAll(descriptorDirectory, fixtureDirPerm); err != nil {
		testRun.Fatal(err)
	}
	if err := os.Symlink(socketLinkTemplate, filepath.Join(descriptorDirectory, "3")); err != nil {
		testRun.Fatal(err)
	}
	if err := os.Symlink(unrelatedLinkPath, filepath.Join(descriptorDirectory, "0")); err != nil {
		testRun.Fatal(err)
	}
	commPath := filepath.Join(procRoot, sshdProcessID, string(commFileName))
	if err := os.WriteFile(commPath, []byte(sshdProcessName+"\n"), fixtureFilePerm); err != nil {
		testRun.Fatal(err)
	}
	// A process directory without a readable fd directory must be tolerated.
	if err := os.MkdirAll(filepath.Join(procRoot, unreadableProcess), fixtureDirPerm); err != nil {
		testRun.Fatal(err)
	}

	listedPorts, err := New(procRoot).ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	namedPorts := 0
	for _, listedPort := range listedPorts {
		if listedPort.Process == "" {
			continue
		}
		namedPorts++
		if listedPort.Process != sshdProcessName || listedPort.Number != sshdPortNumber {
			testRun.Errorf("unexpected named port %+v", listedPort)
		}
	}
	if namedPorts != 1 {
		testRun.Fatalf("named ports = %d, want 1", namedPorts)
	}
}

func TestListPortsContextCancelled(testRun *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New(copyFixtures(testRun)).ListPorts(ctx); err == nil {
		testRun.Fatal("expected context error")
	}
}

func TestDecodeAddress(testRun *testing.T) {
	validCases := []struct {
		encoded  string
		expected string
	}{
		{"00000000", "0.0.0.0"},
		{"0100007F", "127.0.0.1"},
		{"0101A8C0", "192.168.1.1"},
		{"00000000000000000000000000000000", "::"},
		{"00000000000000000000000001000000", "::1"},
		{"B80D0120000000000000000001000000", "2001:db8::1"},
	}
	for _, testCase := range validCases {
		decoded, err := decodeAddress(testCase.encoded)
		if err != nil {
			testRun.Errorf("decodeAddress(%q): %v", testCase.encoded, err)
			continue
		}
		if decoded != testCase.expected {
			testRun.Errorf("decodeAddress(%q) = %q, want %q", testCase.encoded, decoded, testCase.expected)
		}
	}
	for _, invalidEncoded := range []string{"", "ZZZZZZZZ", "0100"} {
		if _, err := decodeAddress(invalidEncoded); err == nil {
			testRun.Errorf("decodeAddress(%q): expected error", invalidEncoded)
		}
	}
}
