package app

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

func makeOutboundTestConnection(remoteAddress string) domain.Connection {
	connection := makeTestConnection(54000, remoteAddress, 443)
	connection.Direction = domain.ConnectionOutbound
	return connection
}

func TestSummarizeConnectionsEmpty(testRun *testing.T) {
	summary := SummarizeConnections(nil)
	if summary.Total != 0 || summary.Inbound != 0 || summary.Outbound != 0 {
		testRun.Fatalf("summary = %+v, want zero counts", summary)
	}
	if len(summary.PerLocalPort) != 0 || len(summary.TopRemoteAddresses) != 0 {
		testRun.Fatalf("summary = %+v, want empty lists", summary)
	}
}

func TestSummarizeConnectionsCountsAndOrdering(testRun *testing.T) {
	connections := []domain.Connection{
		makeTestConnection(443, "198.51.100.7", 50001),
		makeTestConnection(22, "203.0.113.10", 50000),
		makeTestConnection(443, "203.0.113.10", 50002),
		makeTestConnection(80, "192.0.2.55", 50003),
		makeTestConnection(443, "203.0.113.10", 50004),
		makeOutboundTestConnection("93.184.216.34"),
	}
	summary := SummarizeConnections(connections)
	if summary.Total != 6 || summary.Inbound != 5 || summary.Outbound != 1 {
		testRun.Fatalf("counts = %d/%d/%d, want 6/5/1", summary.Total, summary.Inbound, summary.Outbound)
	}
	wantPorts := []PortConnectionCount{{LocalPort: 443, Count: 3}, {LocalPort: 22, Count: 1}, {LocalPort: 80, Count: 1}}
	if !reflect.DeepEqual(summary.PerLocalPort, wantPorts) {
		testRun.Errorf("per port = %v, want %v", summary.PerLocalPort, wantPorts)
	}
	wantRemotes := []RemoteAddressCount{
		{Address: "203.0.113.10", Count: 3},
		{Address: "192.0.2.55", Count: 1},
		{Address: "198.51.100.7", Count: 1},
	}
	if !reflect.DeepEqual(summary.TopRemoteAddresses, wantRemotes) {
		testRun.Errorf("top remotes = %v, want %v", summary.TopRemoteAddresses, wantRemotes)
	}
}

func TestSummarizeConnectionsTopRemotesAreBounded(testRun *testing.T) {
	var connections []domain.Connection
	for peerIndex := 0; peerIndex < maxTopRemoteAddresses+5; peerIndex++ {
		connections = append(connections, makeTestConnection(22, fmt.Sprintf("203.0.113.%d", peerIndex+1), 50000+peerIndex))
	}
	summary := SummarizeConnections(connections)
	if len(summary.TopRemoteAddresses) != maxTopRemoteAddresses {
		testRun.Fatalf("top remotes = %d, want %d", len(summary.TopRemoteAddresses), maxTopRemoteAddresses)
	}
}
