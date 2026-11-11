package baselinefile

import (
	"errors"
	"testing"
)

func TestDefaultPathResolution(testRun *testing.T) {
	home := func() (string, error) { return "/home/alice", nil }
	noHome := func() (string, error) { return "", errors.New("no home") }
	getenv := func(values map[string]string) func(string) string {
		return func(name string) string { return values[name] }
	}
	testCases := []struct {
		name        string
		environment Environment
		want        string
		wantErr     bool
	}{
		{"root", Environment{UserID: 0, Getenv: getenv(nil), HomeDirectory: home}, "/var/lib/exposedwatch/baseline.json", false},
		{"root ignores XDG", Environment{UserID: 0, Getenv: getenv(map[string]string{"XDG_STATE_HOME": "/x"}), HomeDirectory: home}, "/var/lib/exposedwatch/baseline.json", false},
		{"user with absolute XDG", Environment{UserID: 1000, Getenv: getenv(map[string]string{"XDG_STATE_HOME": "/srv/state"}), HomeDirectory: home}, "/srv/state/exposedwatch/baseline.json", false},
		{"user with relative XDG is ignored", Environment{UserID: 1000, Getenv: getenv(map[string]string{"XDG_STATE_HOME": "state"}), HomeDirectory: home}, "/home/alice/.local/state/exposedwatch/baseline.json", false},
		{"user without XDG", Environment{UserID: 1000, Getenv: getenv(nil), HomeDirectory: home}, "/home/alice/.local/state/exposedwatch/baseline.json", false},
		{"user without getenv", Environment{UserID: 1000, HomeDirectory: home}, "/home/alice/.local/state/exposedwatch/baseline.json", false},
		{"user without home", Environment{UserID: 1000, Getenv: getenv(nil), HomeDirectory: noHome}, "", true},
		{"user without home function", Environment{UserID: 1000, Getenv: getenv(nil)}, "", true},
	}
	for _, testCase := range testCases {
		testRun.Run(testCase.name, func(testRun *testing.T) {
			got, err := DefaultPath(testCase.environment)
			if (err != nil) != testCase.wantErr || string(got) != testCase.want {
				testRun.Fatalf("got %q err=%v", got, err)
			}
		})
	}
}
