package procnet

import (
	"context"
	"errors"
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

// connectionTables are the tables that hold TCP connections.
var connectionTables = []socketTable{
	{tcpTableName, domain.TCP, true},
	{tcp6TableName, domain.TCP, true},
}

// ListConnections returns ESTABLISHED TCP connections. A connection is
// inbound when its local port is a LISTEN port of this host (read in the same
// pass), otherwise outbound. Missing tables are skipped; it fails only when
// none could be read. When more than the connection limit exist, the first
// connections are returned together with ports.ErrConnectionsTruncated.
func (reader *Reader) ListConnections(ctx context.Context) ([]domain.Connection, error) {
	listeningPorts := map[int]bool{}
	var establishedRows []connectionRow
	truncated := false
	var tableErrors []error
	anyTableRead := false
	for _, table := range connectionTables {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		err := reader.visitConnectionRows(table, func(row connectionRow) {
			switch row.state {
			case SocketStateListen:
				listeningPorts[row.localPort] = true
			case SocketStateEstablished:
				if len(establishedRows) >= reader.connectionLimit {
					truncated = true
					return
				}
				establishedRows = append(establishedRows, row)
			}
		})
		if err != nil {
			tableErrors = append(tableErrors, err)
			continue
		}
		anyTableRead = true
	}
	if !anyTableRead {
		return nil, fmt.Errorf(noTableReadableMessage, errors.Join(tableErrors...))
	}

	var processNames map[SocketInode]ProcessName
	if len(establishedRows) > 0 {
		processNames = reader.processNamesByInode()
	}
	connections := make([]domain.Connection, 0, len(establishedRows))
	for _, row := range establishedRows {
		direction := domain.ConnectionOutbound
		if listeningPorts[row.localPort] {
			direction = domain.ConnectionInbound
		}
		connections = append(connections, domain.Connection{
			Protocol:      domain.TCP,
			LocalAddress:  row.localAddress,
			LocalPort:     row.localPort,
			RemoteAddress: row.remoteAddress,
			RemotePort:    row.remotePort,
			State:         domain.ConnectionEstablished,
			Direction:     direction,
			ProcessName:   string(processNames[row.inode]),
		})
	}
	if truncated {
		return connections, ports.ErrConnectionsTruncated
	}
	return connections, nil
}
