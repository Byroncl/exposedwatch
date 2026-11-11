package main

import (
	"io"
	"os"
)

// isTerminal reports whether output is an interactive terminal.
func isTerminal(output io.Writer) bool {
	outputFile, isFile := output.(*os.File)
	if !isFile {
		return false
	}
	fileInfo, err := outputFile.Stat()
	return err == nil && fileInfo.Mode()&os.ModeCharDevice != 0
}
