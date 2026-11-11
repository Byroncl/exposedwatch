package main

import (
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// newTestAlert builds the alert of notify-test and of the tui test message.
func newTestAlert(hostName string) domain.Alert {
	return domain.Alert{
		Kind:     domain.AlertKindTest,
		Severity: domain.Info,
		Title:    fmt.Sprintf(notifyTestAlertTitleFormat, hostName),
		Body:     notifyBodyText,
		Host:     hostName,
	}
}
