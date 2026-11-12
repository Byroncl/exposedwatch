package domain

import "fmt"

// AlertKind tells why an alert was raised.
type AlertKind string

// String returns the stable kind name.
func (kind AlertKind) String() string {
	return string(kind)
}

// Alert is a notification about LOCAL exposure. It deliberately has no field
// for remote IP addresses: alerts describe this host, never its peers.
type Alert struct {
	Kind     AlertKind
	Severity Severity
	Title    string
	// Body is the recommendation for a new finding, the finding list for a
	// summary, or the free text of a test alert.
	Body     string
	Host     string
	Port     int // zero when the alert is not about one port
	Protocol Protocol
	Exposure Exposure
	Process  string
	RuleID   string
}

// HasPort reports whether the alert is about one specific port.
func (alert Alert) HasPort() bool {
	return alert.Port > 0
}

// PortLabel returns "port/protocol", or an empty string without a port.
func (alert Alert) PortLabel() string {
	if !alert.HasPort() {
		return ""
	}
	return fmt.Sprintf(portLabelFormat, alert.Port, alert.Protocol)
}
