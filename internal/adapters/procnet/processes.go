package procnet

import (
	"os"
	"path/filepath"
	"strings"
)

// ProcessName is a process command name read from /proc/<pid>/comm.
type ProcessName string

// processNamesByInode maps socket inodes to process names, best effort:
// unreadable processes are skipped and their sockets keep an empty name.
func (reader *Reader) processNamesByInode() map[SocketInode]ProcessName {
	processNames := map[SocketInode]ProcessName{}
	processDirectories, err := filepath.Glob(filepath.Join(reader.procRoot, processDirectoryGlob))
	if err != nil {
		return processNames
	}
	for _, processDirectory := range processDirectories {
		descriptorDirectory := filepath.Join(processDirectory, string(fdDirectoryName))
		descriptors, err := os.ReadDir(descriptorDirectory)
		if err != nil {
			continue
		}
		var processName ProcessName
		for _, descriptor := range descriptors {
			linkTarget, err := os.Readlink(filepath.Join(descriptorDirectory, descriptor.Name()))
			if err != nil || !strings.HasPrefix(linkTarget, socketLinkPrefix) || !strings.HasSuffix(linkTarget, socketLinkSuffix) {
				continue
			}
			if processName == "" {
				commContent, err := os.ReadFile(filepath.Join(processDirectory, string(commFileName)))
				if err != nil {
					break
				}
				processName = ProcessName(strings.TrimSpace(string(commContent)))
			}
			socketInode := SocketInode(strings.TrimSuffix(strings.TrimPrefix(linkTarget, socketLinkPrefix), socketLinkSuffix))
			if _, alreadyOwned := processNames[socketInode]; !alreadyOwned {
				processNames[socketInode] = processName
			}
		}
	}
	return processNames
}
