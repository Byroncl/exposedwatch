package dockerapi

import (
	"context"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// ListPorts returns every published container port.
func (client *Client) ListPorts(ctx context.Context) ([]domain.Port, error) {
	containers, err := client.listContainers(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPorts(containers), nil
}
