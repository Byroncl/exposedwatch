package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Documentation-range and private addresses only: no real peers in fixtures.
const (
	testServerAddress   = "10.0.0.5"
	testLoopbackAddress = "127.0.0.1"
	testSshPeer         = "203.0.113.7"
	testWebPeer         = "198.51.100.23"
	testCurlPeer        = "192.0.2.10"
	testDatabasePeer    = "192.0.2.53"
)

var errConnectionsDown = errors.New("cannot read the connection tables")

type fakeConnectionSource struct {
	snapshot app.ConnectionSnapshot
	err      error
}

func (source fakeConnectionSource) Snapshot(_ context.Context) (app.ConnectionSnapshot, error) {
	return source.snapshot, source.err
}

func testConnection(direction domain.ConnectionDirection, localPort int, remoteAddress string, remotePort int, process string) domain.Connection {
	localAddress := testServerAddress
	if remoteAddress == testLoopbackAddress {
		localAddress = testLoopbackAddress
	}
	return domain.Connection{
		Protocol: domain.TCP, LocalAddress: localAddress, LocalPort: localPort,
		RemoteAddress: remoteAddress, RemotePort: remotePort,
		State: domain.ConnectionEstablished, Direction: direction, ProcessName: process,
	}
}

func connectionSshInbound() domain.Connection {
	return testConnection(domain.ConnectionInbound, 22, testSshPeer, 51000, "sshd")
}

func connectionWebInbound() domain.Connection {
	return testConnection(domain.ConnectionInbound, 443, testWebPeer, 40001, "nginx")
}

func connectionMetricsInbound() domain.Connection {
	return testConnection(domain.ConnectionInbound, 9000, testSshPeer, 40002, "metrics")
}

func connectionCurlOutbound() domain.Connection {
	return testConnection(domain.ConnectionOutbound, 4000, testCurlPeer, 443, "curl")
}

func connectionDatabaseOutbound() domain.Connection {
	return testConnection(domain.ConnectionOutbound, 50000, testDatabasePeer, 5432, "psql")
}

func connectionLoopbackInbound() domain.Connection {
	return testConnection(domain.ConnectionInbound, 5432, testLoopbackAddress, 50210, "postgres")
}

// baselineConnections is the first snapshot of most tests.
func baselineConnections() []domain.Connection {
	return []domain.Connection{connectionSshInbound(), connectionWebInbound(), connectionCurlOutbound()}
}

// fullConnections is the baseline plus one new inbound, one new outbound and
// one loopback-to-loopback connection.
func fullConnections() []domain.Connection {
	return append(baselineConnections(), connectionMetricsInbound(), connectionDatabaseOutbound(), connectionLoopbackInbound())
}

func snapshotOf(connections []domain.Connection) app.ConnectionSnapshot {
	return app.ConnectionSnapshot{Connections: connections, Summary: app.SummarizeConnections(connections)}
}

func connectionsResult(connections []domain.Connection, offsetSeconds int) ConnectionsResultMessage {
	return ConnectionsResultMessage{
		Snapshot: snapshotOf(connections),
		At:       testRefreshTime.Add(time.Duration(offsetSeconds) * time.Second),
	}
}

// newConnectionsModel builds a model with connections enabled, sized, with
// the first connection snapshot as the baseline, on the Connections tab.
func newConnectionsModel(testRun *testing.T, connections []domain.Connection) Model {
	testRun.Helper()
	model := New(Options{
		Scanner: fakeScanner{}, Connections: fakeConnectionSource{}, Interval: testInterval, Version: testVersion,
	})
	model = applyMessage(testRun, model, tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	model = applyMessage(testRun, model, RefreshResultMessage{Report: sampleReport(), At: testRefreshTime})
	model = applyMessage(testRun, model, connectionsResult(connections, 0))
	return typeKeys(testRun, model, "4")
}

func observeConnections(testRun *testing.T, model Model, connections []domain.Connection, offsetSeconds int) Model {
	testRun.Helper()
	return applyMessage(testRun, model, connectionsResult(connections, offsetSeconds))
}

func rowForLocalPort(testRun *testing.T, model Model, localPort int) displayRow {
	testRun.Helper()
	for _, row := range model.visibleRows {
		if row.connection.LocalPort == localPort {
			return row
		}
	}
	testRun.Fatalf("no row with local port %d among %v", localPort, visibleLocalPorts(model))
	return displayRow{}
}

func visibleLocalPorts(model Model) []int {
	numbers := make([]int, 0, len(model.visibleRows))
	for _, row := range model.visibleRows {
		numbers = append(numbers, row.connection.LocalPort)
	}
	return numbers
}
