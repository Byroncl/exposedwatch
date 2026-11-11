package main

import (
	"os"
	"path/filepath"
	"testing"
)

const fakeTcpTable = `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1 0000000000000000 100 0 0 10 0
`

func writeFakeProc(testRun *testing.T, procRoot string) {
	testRun.Helper()
	netDirectory := filepath.Join(procRoot, "net")
	if err := os.MkdirAll(netDirectory, 0o755); err != nil {
		testRun.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(netDirectory, "tcp"), []byte(fakeTcpTable), 0o644); err != nil {
		testRun.Fatal(err)
	}
}
