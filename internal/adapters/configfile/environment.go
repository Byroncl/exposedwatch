package configfile

import (
	"errors"
	"os"
	"path/filepath"
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

// DefaultPath resolves where the config file lives: root uses /etc, other
// users use $XDG_CONFIG_HOME or ~/.config.
func DefaultPath(environment Environment) (string, error) {
	if environment.UserID == rootUserID {
		return rootConfigPath, nil
	}
	if environment.Getenv != nil {
		configHome := environment.Getenv(envXDGConfigHome)
		if filepath.IsAbs(configHome) {
			return filepath.Join(configHome, applicationFolder, configFileName), nil
		}
	}
	if environment.HomeDirectory == nil {
		return "", errors.New(noHomeMessage)
	}
	home, err := environment.HomeDirectory()
	if err != nil || home == "" {
		return "", errors.New(noHomeMessage)
	}
	return filepath.Join(home, homeConfigDirectory, applicationFolder, configFileName), nil
}
