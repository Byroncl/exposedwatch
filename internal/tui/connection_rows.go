package tui

import (
	"net"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// connectionsPassingLoopbackToggle lists the snapshot connections that the
// loopback toggle lets through. Summaries and counts start from this set.
func (model Model) connectionsPassingLoopbackToggle() []domain.Connection {
	all := model.connectionSnapshot.Connections
	if !model.hideLoopback {
		return all
	}
	kept := make([]domain.Connection, 0, len(all))
	for _, connection := range all {
		if !connection.IsLoopbackToLoopback() {
			kept = append(kept, connection)
		}
	}
	return kept
}

// connectionRows builds the unsorted rows of the Connections tab: the
// connections that pass the loopback toggle and the text filter.
func (model Model) connectionRows() []displayRow {
	connections := model.connectionsPassingLoopbackToggle()
	rows := make([]displayRow, 0, len(connections))
	for _, connection := range connections {
		if !connectionMatchesText(connection, model.connectionFilterText) {
			continue
		}
		identity := app.ConnectionIdentityOf(connection)
		row := displayRow{
			key:          rowKey{connection: identity},
			isConnection: true,
			connection:   connection,
			isNew:        model.connectionTracker.isNew(identity),
		}
		if seenAt, known := model.connectionTracker.firstSeenAt(identity); known {
			row.connectionAge = model.connectionsAt.Sub(seenAt)
		}
		rows = append(rows, row)
	}
	return rows
}

// connectionMatchesText reports whether the remote address, remote port or
// process name contains the text, ignoring case.
func connectionMatchesText(connection domain.Connection, text string) bool {
	if text == "" {
		return true
	}
	fields := []string{
		connection.RemoteAddress,
		strconv.Itoa(connection.RemotePort),
		connection.RemoteAddress + ":" + strconv.Itoa(connection.RemotePort),
		connection.ProcessName,
	}
	return strings.Contains(strings.ToLower(strings.Join(fields, " ")), strings.ToLower(text))
}

// remoteEndpoint renders the remote side as address:port (IPv6 in brackets).
func remoteEndpoint(connection domain.Connection) string {
	return net.JoinHostPort(connection.RemoteAddress, strconv.Itoa(connection.RemotePort))
}

// sortConnectionRows orders connection rows in place. Every mode falls back to
// the default order: inbound first, then local port, then remote address.
func sortConnectionRows(rows []displayRow, mode ConnectionSortMode) {
	sort.SliceStable(rows, func(left, right int) bool {
		return connectionRowLess(rows[left], rows[right], mode)
	})
}

func connectionRowLess(left, right displayRow, mode ConnectionSortMode) bool {
	switch mode {
	case ConnectionSortByLocalPort:
		if left.connection.LocalPort != right.connection.LocalPort {
			return left.connection.LocalPort < right.connection.LocalPort
		}
	case ConnectionSortByRemoteAddress:
		if order := compareRemoteAddresses(left.connection.RemoteAddress, right.connection.RemoteAddress); order != 0 {
			return order < 0
		}
	case ConnectionSortByProcess:
		leftName, rightName := strings.ToLower(left.connection.ProcessName), strings.ToLower(right.connection.ProcessName)
		if leftName != rightName {
			// Connections without a known process go last.
			return rightName == "" || (leftName != "" && leftName < rightName)
		}
	case ConnectionSortByAge:
		if left.connectionAge != right.connectionAge {
			return left.connectionAge < right.connectionAge // newest first
		}
	}
	return connectionDefaultLess(left, right)
}

// connectionDefaultLess is the default order: inbound first, then local port
// ascending, then remote address, then remote port.
func connectionDefaultLess(left, right displayRow) bool {
	leftConnection, rightConnection := left.connection, right.connection
	if leftConnection.Direction != rightConnection.Direction {
		return leftConnection.Direction < rightConnection.Direction
	}
	if leftConnection.LocalPort != rightConnection.LocalPort {
		return leftConnection.LocalPort < rightConnection.LocalPort
	}
	if order := compareRemoteAddresses(leftConnection.RemoteAddress, rightConnection.RemoteAddress); order != 0 {
		return order < 0
	}
	return leftConnection.RemotePort < rightConnection.RemotePort
}

// compareRemoteAddresses orders IP addresses numerically and falls back to
// text for anything that does not parse.
func compareRemoteAddresses(left, right string) int {
	leftAddress, leftErr := netip.ParseAddr(left)
	rightAddress, rightErr := netip.ParseAddr(right)
	if leftErr == nil && rightErr == nil {
		return leftAddress.Compare(rightAddress)
	}
	return strings.Compare(left, right)
}
