package baselinefile

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Byroncl/exposedwatch/internal/ports"
)

// Environment is everything DefaultPath needs from the process, injectable
// for tests.
type Environment struct {
	UserID        int
	Getenv        func(name string) string
	HomeDirectory func() (string, error)
}

// SystemEnvironment reads the real process environment.
func SystemEnvironment() Environment {
	return Environment{UserID: os.Geteuid(), Getenv: os.Getenv, HomeDirectory: os.UserHomeDir}
}

// DefaultPath resolves where the baseline lives: root uses /var/lib, other
// users use $XDG_STATE_HOME (when absolute) or ~/.local/state.
func DefaultPath(environment Environment) (ports.BaselinePath, error) {
	if environment.UserID == rootUserID {
		return rootBaselinePath, nil
	}
	if environment.Getenv != nil {
		stateHome := environment.Getenv(envXDGStateHome)
		if filepath.IsAbs(stateHome) {
			return ports.BaselinePath(filepath.Join(stateHome, applicationFolder, baselineFileName)), nil
		}
	}
	if environment.HomeDirectory == nil {
		return "", errors.New(noHomeMessage)
	}
	home, err := environment.HomeDirectory()
	if err != nil || home == "" {
		return "", errors.New(noHomeMessage)
	}
	return ports.BaselinePath(filepath.Join(home, homeStateDirectory, applicationFolder, baselineFileName)), nil
}
