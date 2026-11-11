package tui

import "fmt"

// loopbackState names whether loopback-to-loopback connections are hidden.
func (model Model) loopbackState() string {
	if model.hideLoopback {
		return loopbackStateHidden
	}
	return loopbackStateShown
}

// loopbackHintText is the footer hint of the loopback toggle.
func loopbackHintText(model Model) string {
	return fmt.Sprintf(loopbackFooterFormat, model.loopbackState())
}

// connectionWarningLines lists the non-fatal connection problems. They only
// show on the Connections tab, where they matter.
func (model Model) connectionWarningLines() []string {
	if model.activeTab != TabConnections || !model.connectionsEnabled() {
		return nil
	}
	var lines []string
	if model.connectionsErr != nil {
		lines = append(lines, fmt.Sprintf(connectionsWarningFormat, model.connectionsErr))
	}
	if model.connectionSnapshot.Truncated {
		lines = append(lines, fmt.Sprintf(connectionsTruncatedFormat, len(model.connectionSnapshot.Connections)))
	}
	return lines
}

// connectionsEmptyMessage explains why the Connections table has no rows.
func (model Model) connectionsEmptyMessage() string {
	switch {
	case !model.connectionsEnabled():
		return connectionsDisabledMessage
	case !model.hasConnections && model.connectionsErr != nil:
		return connectionsUnavailableText
	case !model.hasConnections:
		return connectionsReadingMessage
	case model.connectionFilterText != "":
		return noMatchMessage
	case model.hideLoopback && len(model.connectionSnapshot.Connections) > 0:
		return loopbackHiddenMessage
	default:
		return noConnectionsMessage
	}
}
