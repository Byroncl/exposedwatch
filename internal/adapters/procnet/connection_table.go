package procnet

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// connectionRow is one parsed TCP row with both endpoints decoded.
type connectionRow struct {
	state         SocketState
	localAddress  string
	localPort     int
	remoteAddress string
	remotePort    int
	inode         SocketInode
}

// visitConnectionRows calls visit for every well-formed row of a TCP table.
// Malformed rows are skipped. Rows are streamed so memory stays flat.
func (reader *Reader) visitConnectionRows(table socketTable, visit func(connectionRow)) error {
	tablePath := filepath.Join(reader.procRoot, string(netDirectoryName), string(table.fileName))
	tableFile, err := os.Open(tablePath)
	if err != nil {
		return err
	}
	defer tableFile.Close()

	rowScanner := bufio.NewScanner(tableFile)
	rowScanner.Scan() // header row
	for rowScanner.Scan() {
		if row, parsed := parseConnectionRow(rowScanner.Text()); parsed {
			visit(row)
		}
	}
	return rowScanner.Err()
}

func parseConnectionRow(row string) (connectionRow, bool) {
	columns := strings.Fields(row)
	if len(columns) < minimumColumnCount {
		return connectionRow{}, false
	}
	localAddress, localPort, err := decodeEndpoint(columns[localAddressColumnIndex])
	if err != nil {
		return connectionRow{}, false
	}
	remoteAddress, remotePort, err := decodeEndpoint(columns[remoteAddressColumnIndex])
	if err != nil {
		return connectionRow{}, false
	}
	return connectionRow{
		state:         SocketState(columns[stateColumnIndex]),
		localAddress:  localAddress,
		localPort:     localPort,
		remoteAddress: remoteAddress,
		remotePort:    remotePort,
		inode:         SocketInode(columns[inodeColumnIndex]),
	}, true
}
