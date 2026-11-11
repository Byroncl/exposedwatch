package procnet

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

var _ ports.ConnectionSource = (*Reader)(nil)

const (
	connectionFixtureDirectory = "testdata/connections"
	fixtureEstablishedCount    = 9
	fixtureInboundCount        = 7
	fixtureOutboundCount       = 2
	truncationTestLimit        = 3
	sshInboundInode            = "2001"
	sshInboundProcessName      = "sshd"
	sshInboundLocalPort        = 22
	mappedRemoteAddress        = "203.0.113.10"
	mappedLocalAddress         = "10.0.0.5"
	ipv6LocalAddress           = "2001:db8::1"
	ipv6RemoteAddress          = "2001:db8::2"
	publicOutboundRemote       = "93.184.216.34"
)

var connectionFixtureTables = []ProcFileName{tcpTableName, tcp6TableName}

func copyConnectionFixtures(testRun *testing.T) string {
	testRun.Helper()
	procRoot := testRun.TempDir()
	if err := os.MkdirAll(filepath.Join(procRoot, string(netDirectoryName)), fixtureDirPerm); err != nil {
		testRun.Fatal(err)
	}
	for _, tableName := range connectionFixtureTables {
		relativePath := filepath.Join(string(netDirectoryName), string(tableName))
		content, err := os.ReadFile(filepath.Join(connectionFixtureDirectory, relativePath))
		if err != nil {
			testRun.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(procRoot, relativePath), content, fixtureFilePerm); err != nil {
			testRun.Fatal(err)
		}
	}
	return procRoot
}

func findConnection(connections []domain.Connection, localPort int, remoteAddress string) (domain.Connection, bool) {
	for _, connection := range connections {
		if connection.LocalPort == localPort && connection.RemoteAddress == remoteAddress {
			return connection, true
		}
	}
	return domain.Connection{}, false
}

func TestListConnectionsReadsOnlyEstablished(testRun *testing.T) {
	connections, err := New(copyConnectionFixtures(testRun)).ListConnections(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if len(connections) != fixtureEstablishedCount {
		testRun.Fatalf("got %d connections, want %d: %+v", len(connections), fixtureEstablishedCount, connections)
	}
	inboundCount, outboundCount := 0, 0
	for _, connection := range connections {
		if connection.State != domain.ConnectionEstablished || connection.Protocol != domain.TCP {
			testRun.Errorf("unexpected state or protocol: %+v", connection)
		}
		if connection.Direction == domain.ConnectionInbound {
			inboundCount++
		} else {
			outboundCount++
		}
	}
	if inboundCount != fixtureInboundCount || outboundCount != fixtureOutboundCount {
		testRun.Errorf("inbound/outbound = %d/%d, want %d/%d", inboundCount, outboundCount, fixtureInboundCount, fixtureOutboundCount)
	}
}

func TestListConnectionsDecodesEndpointsAndDirection(testRun *testing.T) {
	connections, err := New(copyConnectionFixtures(testRun)).ListConnections(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	inboundSSH, found := findConnection(connections, sshInboundLocalPort, mappedRemoteAddress)
	if !found {
		testRun.Fatal("inbound ssh connection missing")
	}
	if inboundSSH.LocalAddress != mappedLocalAddress || inboundSSH.RemotePort != 50000 || inboundSSH.Direction != domain.ConnectionInbound {
		testRun.Errorf("unexpected ssh connection: %+v", inboundSSH)
	}
	outbound, found := findConnection(connections, 54000, publicOutboundRemote)
	if !found || outbound.Direction != domain.ConnectionOutbound || outbound.RemotePort != 443 {
		testRun.Errorf("unexpected outbound connection: %+v (found=%v)", outbound, found)
	}
}

func TestListConnectionsLoopbackPostgresBothSides(testRun *testing.T) {
	connections, err := New(copyConnectionFixtures(testRun)).ListConnections(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	loopbackCount := 0
	for _, connection := range connections {
		if connection.IsLoopbackToLoopback() {
			loopbackCount++
		}
	}
	if loopbackCount != 2 {
		testRun.Fatalf("loopback connections = %d, want 2", loopbackCount)
	}
	serverSide, found := findConnection(connections, 5432, "127.0.0.1")
	if !found || serverSide.Direction != domain.ConnectionInbound {
		testRun.Errorf("postgres server side: %+v (found=%v)", serverSide, found)
	}
	clientSide, found := findConnection(connections, 40001, "127.0.0.1")
	if !found || clientSide.Direction != domain.ConnectionOutbound {
		testRun.Errorf("postgres client side: %+v (found=%v)", clientSide, found)
	}
}

func TestListConnectionsIPv6AndMappedAddresses(testRun *testing.T) {
	connections, err := New(copyConnectionFixtures(testRun)).ListConnections(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	ipv6Connection, found := findConnection(connections, 8080, ipv6RemoteAddress)
	if !found || ipv6Connection.LocalAddress != ipv6LocalAddress || ipv6Connection.Direction != domain.ConnectionInbound {
		testRun.Errorf("ipv6 connection: %+v (found=%v)", ipv6Connection, found)
	}
	mappedConnection, found := findConnection(connections, 443, mappedRemoteAddress)
	if !found {
		testRun.Fatal("mapped connection missing")
	}
	// Two 443 connections share the remote IP (one IPv4, one mapped IPv6);
	// both must be normalized to plain IPv4.
	for _, connection := range connections {
		if connection.LocalAddress == "::ffff:"+mappedLocalAddress || connection.RemoteAddress == "::ffff:"+mappedRemoteAddress {
			testRun.Errorf("address not normalized: %+v", connection)
		}
	}
	if mappedConnection.LocalAddress != mappedLocalAddress {
		testRun.Errorf("mapped local = %q, want %q", mappedConnection.LocalAddress, mappedLocalAddress)
	}
}

func TestListConnectionsTruncates(testRun *testing.T) {
	reader := New(copyConnectionFixtures(testRun))
	reader.connectionLimit = truncationTestLimit
	connections, err := reader.ListConnections(context.Background())
	if !errors.Is(err, ports.ErrConnectionsTruncated) {
		testRun.Fatalf("err = %v, want ErrConnectionsTruncated", err)
	}
	if len(connections) != truncationTestLimit {
		testRun.Fatalf("got %d connections, want %d", len(connections), truncationTestLimit)
	}
}

func TestListConnectionsAtLimitIsNotTruncated(testRun *testing.T) {
	reader := New(copyConnectionFixtures(testRun))
	reader.connectionLimit = fixtureEstablishedCount
	connections, err := reader.ListConnections(context.Background())
	if err != nil || len(connections) != fixtureEstablishedCount {
		testRun.Fatalf("got %d connections, err %v", len(connections), err)
	}
}

func TestListConnectionsMissingTableTolerated(testRun *testing.T) {
	procRoot := copyConnectionFixtures(testRun)
	if err := os.Remove(filepath.Join(procRoot, string(netDirectoryName), string(tcp6TableName))); err != nil {
		testRun.Fatal(err)
	}
	connections, err := New(procRoot).ListConnections(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if len(connections) != fixtureEstablishedCount-2 {
		testRun.Fatalf("got %d connections, want %d", len(connections), fixtureEstablishedCount-2)
	}
}

func TestListConnectionsNoTableIsError(testRun *testing.T) {
	if _, err := New(testRun.TempDir()).ListConnections(context.Background()); err == nil {
		testRun.Fatal("expected error when no table is readable")
	}
}

func TestListConnectionsContextCancelled(testRun *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New(copyConnectionFixtures(testRun)).ListConnections(ctx); err == nil {
		testRun.Fatal("expected context error")
	}
}

func TestListConnectionsProcessName(testRun *testing.T) {
	procRoot := copyConnectionFixtures(testRun)
	descriptorDirectory := filepath.Join(procRoot, sshdProcessID, string(fdDirectoryName))
	if err := os.MkdirAll(descriptorDirectory, fixtureDirPerm); err != nil {
		testRun.Fatal(err)
	}
	socketLink := socketLinkPrefix + sshInboundInode + socketLinkSuffix
	if err := os.Symlink(socketLink, filepath.Join(descriptorDirectory, "3")); err != nil {
		testRun.Fatal(err)
	}
	commPath := filepath.Join(procRoot, sshdProcessID, string(commFileName))
	if err := os.WriteFile(commPath, []byte(sshInboundProcessName+"\n"), fixtureFilePerm); err != nil {
		testRun.Fatal(err)
	}
	connections, err := New(procRoot).ListConnections(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	namedCount := 0
	for _, connection := range connections {
		if connection.ProcessName == "" {
			continue
		}
		namedCount++
		if connection.ProcessName != sshInboundProcessName || connection.LocalPort != sshInboundLocalPort {
			testRun.Errorf("unexpected named connection: %+v", connection)
		}
	}
	if namedCount != 1 {
		testRun.Fatalf("named connections = %d, want 1", namedCount)
	}
}
