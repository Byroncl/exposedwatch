package procnet

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// socketTable describes one /proc/net table.
type socketTable struct {
	fileName            ProcFileName
	protocol            domain.Protocol
	requiresListenState bool
}

var socketTables = []socketTable{
	{tcpTableName, domain.TCP, true},
	{tcp6TableName, domain.TCP, true},
	{udpTableName, domain.UDP, false},
	{udp6TableName, domain.UDP, false},
}

// socketEntry is one parsed row plus its socket inode.
type socketEntry struct {
	port  domain.Port
	inode SocketInode
}

func (reader *Reader) readTable(table socketTable) ([]socketEntry, error) {
	tablePath := filepath.Join(reader.procRoot, string(netDirectoryName), string(table.fileName))
	tableFile, err := os.Open(tablePath)
	if err != nil {
		return nil, err
	}
	defer tableFile.Close()

	var tableEntries []socketEntry
	rowScanner := bufio.NewScanner(tableFile)
	rowScanner.Scan() // header row
	for rowScanner.Scan() {
		if socketEntry, parsed := parseRow(rowScanner.Text(), table); parsed {
			tableEntries = append(tableEntries, socketEntry)
		}
	}
	return tableEntries, rowScanner.Err()
}

// parseRow parses one socket row; malformed or non-listening rows are skipped.
func parseRow(row string, table socketTable) (socketEntry, bool) {
	columns := strings.Fields(row)
	if len(columns) < minimumColumnCount {
		return socketEntry{}, false
	}
	if table.requiresListenState && SocketState(columns[stateColumnIndex]) != SocketStateListen {
		return socketEntry{}, false
	}
	addressHex, portHex, found := strings.Cut(columns[localAddressColumnIndex], addressPortSeparator)
	if !found {
		return socketEntry{}, false
	}
	portNumber, err := strconv.ParseUint(portHex, hexadecimalBase, portBitSize)
	if err != nil || portNumber == unboundPortNumber {
		return socketEntry{}, false
	}
	bindAddress, err := decodeAddress(addressHex)
	if err != nil {
		return socketEntry{}, false
	}
	return socketEntry{
		port: domain.Port{
			Number:      int(portNumber),
			Protocol:    table.protocol,
			BindAddress: bindAddress,
			Source:      domain.SourceHost,
		},
		inode: SocketInode(columns[inodeColumnIndex]),
	}, true
}
