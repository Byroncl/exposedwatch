// Package procnet reads listening sockets from the Linux /proc/net tables.
package procnet

import (
	"context"
	"errors"
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Reader lists host ports from a proc tree.
type Reader struct {
	procRoot        string
	connectionLimit int
}

// New returns a Reader for the given proc root. An empty root means
// DefaultProcRoot.
func New(procRoot string) *Reader {
	if procRoot == "" {
		procRoot = DefaultProcRoot
	}
	return &Reader{procRoot: procRoot, connectionLimit: DefaultConnectionLimit}
}

// ListPorts returns listening TCP sockets and bound UDP sockets. Missing
// tables (for example IPv6 disabled) are skipped; it fails only when none
// could be read.
func (reader *Reader) ListPorts(ctx context.Context) ([]domain.Port, error) {
	var socketEntries []socketEntry
	var tableErrors []error
	anyTableRead := false
	for _, socketTable := range socketTables {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tableEntries, err := reader.readTable(socketTable)
		if err != nil {
			tableErrors = append(tableErrors, err)
			continue
		}
		anyTableRead = true
		socketEntries = append(socketEntries, tableEntries...)
	}
	if !anyTableRead {
		return nil, fmt.Errorf(noTableReadableMessage, errors.Join(tableErrors...))
	}

	processNames := reader.processNamesByInode()
	listeningPorts := make([]domain.Port, 0, len(socketEntries))
	for _, socketEntry := range socketEntries {
		socketEntry.port.Process = string(processNames[socketEntry.inode])
		listeningPorts = append(listeningPorts, socketEntry.port)
	}
	return listeningPorts, nil
}
