package dockerapi

import (
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// publishedPort is one port entry of the Docker container list payload.
type publishedPort struct {
	IP         string `json:"IP"`
	PublicPort int    `json:"PublicPort"`
	Type       string `json:"Type"`
}

// container is the subset of the Docker container payload this adapter uses.
type container struct {
	Names []string        `json:"Names"`
	Ports []publishedPort `json:"Ports"`
}

// toDomainPorts maps host-published container ports into domain ports.
func toDomainPorts(containers []container) []domain.Port {
	var domainPorts []domain.Port
	for _, listedContainer := range containers {
		containerName := ""
		if len(listedContainer.Names) > 0 {
			containerName = strings.TrimPrefix(listedContainer.Names[0], containerNamePrefix)
		}
		for _, published := range listedContainer.Ports {
			if published.PublicPort == unpublishedPortNumber {
				continue
			}
			domainPorts = append(domainPorts, domain.Port{
				Number:      published.PublicPort,
				Protocol:    domain.Protocol(published.Type),
				BindAddress: published.IP,
				Source:      domain.SourceDocker,
				Container:   containerName,
			})
		}
	}
	return domainPorts
}
