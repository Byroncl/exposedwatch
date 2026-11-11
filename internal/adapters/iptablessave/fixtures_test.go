package iptablessave

import (
	"os"
	"path/filepath"
	"testing"
)

const fixtureDirectory = "testdata"

func readFixture(testRun *testing.T, fileName string) []byte {
	testRun.Helper()
	content, err := os.ReadFile(filepath.Join(fixtureDirectory, fileName))
	if err != nil {
		testRun.Fatal(err)
	}
	return content
}
