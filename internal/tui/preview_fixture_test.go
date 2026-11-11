package tui

import (
	"testing"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/risk"
)

const previewNewPortNumber = 6379

func previewPort(number int, protocol domain.Protocol, bindAddress, process, container string, source domain.Source) domain.Port {
	return domain.Port{Number: number, Protocol: protocol, BindAddress: bindAddress, Process: process, Container: container, Source: source}
}

func previewPorts() []domain.Port {
	return []domain.Port{
		previewPort(22, domain.TCP, "0.0.0.0", "sshd", "", domain.SourceHost),
		previewPort(23, domain.TCP, "0.0.0.0", "telnetd", "", domain.SourceHost),
		previewPort(80, domain.TCP, "0.0.0.0", "nginx", "", domain.SourceHost),
		previewPort(443, domain.TCP, "0.0.0.0", "nginx", "", domain.SourceHost),
		previewPort(631, domain.TCP, "127.0.0.1", "cupsd", "", domain.SourceHost),
		previewPort(5432, domain.TCP, "10.0.0.5", "postgres", "", domain.SourceHost),
		previewPort(2375, domain.TCP, "0.0.0.0", "dockerd", "", domain.SourceHost),
		previewPort(6379, domain.TCP, "0.0.0.0", "redis-server", "cache", domain.SourceDocker),
		previewPort(8080, domain.TCP, "0.0.0.0", "", "web-app", domain.SourceDocker),
		previewPort(9000, domain.TCP, "172.17.0.1", "", "metrics", domain.SourceDocker),
		previewPort(53, domain.UDP, "127.0.0.53", "systemd-resolve", "", domain.SourceHost),
		previewPort(8443, domain.TCP, "203.0.113.10", "", "", domain.SourceExternal),
	}
}

func previewReport(ports []domain.Port) app.Report {
	return app.Report{Ports: ports, Findings: risk.Evaluate(ports)}
}

// previewModel builds a model with a baseline, one NEW port and two closed ports.
func previewModel(testRun *testing.T) Model {
	testRun.Helper()
	var baseline []domain.Port
	for _, port := range previewPorts() {
		if port.Number != previewNewPortNumber {
			baseline = append(baseline, port)
		}
	}
	closedFirst := previewPort(3306, domain.TCP, "0.0.0.0", "mysqld", "", domain.SourceHost)
	closedSecond := previewPort(3389, domain.TCP, "0.0.0.0", "xrdp", "", domain.SourceHost)
	model := New(Options{Scanner: fakeScanner{}, Interval: testInterval, Version: testVersion})
	model = refreshWith(testRun, model, previewReport(append(append([]domain.Port{}, baseline...), closedFirst, closedSecond)), 0)
	model = refreshWith(testRun, model, previewReport(append(previewPorts(), closedSecond)), 1)
	return refreshWith(testRun, model, previewReport(previewPorts()), 2)
}
