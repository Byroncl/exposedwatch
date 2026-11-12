package tui

import "github.com/Byroncl/exposedwatch/internal/app"

// connectionsEnabled tells whether a connection source was configured.
func (model Model) connectionsEnabled() bool {
	return model.connections != nil
}

// acceptsSortAndSearch tells whether sort and search keys act on the active
// tab. The Connections tab ignores them while it is disabled.
func (model Model) acceptsSortAndSearch() bool {
	if model.activeTab == TabConnections && !model.connectionsEnabled() {
		return false
	}
	if model.activeTab == TabFirewallPorts && !model.firewallEnabled() {
		return false
	}
	return model.activeTab.supportsSortAndSearch()
}

// handleConnectionsResult stores a connection read. A failed read keeps the
// previous data and never touches the tracker, so it cannot fake closed or
// new connections.
func (model Model) handleConnectionsResult(result ConnectionsResultMessage) Model {
	model.connectionsRefreshing = false
	model.connectionsErr = result.Err
	if result.Err != nil {
		return model
	}
	model.connectionSnapshot = result.Snapshot
	model.hasConnections = true
	model.connectionsAt = result.At
	model.connectionTracker = model.connectionTracker.observe(result.Snapshot.Connections, result.At)
	model.rebuildVisibleRows()
	return model
}

// toggleLoopback shows or hides loopback-to-loopback connections. It only
// acts on the connection views (Live and Top remote).
func (model *Model) toggleLoopback() {
	if !model.activeTab.usesConnections() || !model.connectionsEnabled() {
		return
	}
	model.hideLoopback = !model.hideLoopback
	model.rebuildVisibleRows()
}

// toggleInboundOnly shows only inbound connections, or all of them again. It
// only acts on the connection views (Live and Top remote).
func (model *Model) toggleInboundOnly() {
	if !model.activeTab.usesConnections() || !model.connectionsEnabled() {
		return
	}
	model.inboundOnly = !model.inboundOnly
	model.rebuildVisibleRows()
}

// refreshConnectionCaches recomputes the summary of the connections that pass
// the toggles and the count that also passes the text filter.
func (model *Model) refreshConnectionCaches() {
	connections := model.connectionsPassingToggles()
	model.connectionSummary = app.SummarizeConnections(connections)
	displayed := 0
	for _, connection := range connections {
		if connectionMatchesText(connection, model.connectionFilterText) {
			displayed++
		}
	}
	model.connectionDisplayCount = displayed
}
