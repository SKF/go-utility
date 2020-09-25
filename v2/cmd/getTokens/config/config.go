package config

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"

	"github.com/SKF/go-utility/v2/cmd/getTokens/model"
	"gopkg.in/yaml.v2"
)

func GetConfigDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		err = fmt.Errorf("failed to get current user: %w", err)
		return "", err
	}

	return path.Join(usr.HomeDir, ".skf"), nil
}

func Read(environ string) (cfg model.Config, err error) {
	usr, err := user.Current()
	if err != nil {
		err = fmt.Errorf("failed to get current user: %w", err)
		return
	}

	file, err := os.Open(path.Join(usr.HomeDir, ".skf/config.yaml"))
	if err != nil {
		err = fmt.Errorf("failed to open config: %w", err)
		return
	}
	defer file.Close()

	return ReadFile(file, environ)
}

func ReadFile(file io.Reader, environ string) (cfg model.Config, err error) {
	configs := map[string]model.Config{}

	err = yaml.NewDecoder(file).Decode(&configs)
	if err != nil {
		err = fmt.Errorf("failed to parse config: %w", err)
		return
	}

	cfg, ok := configs[environ]
	if !ok {
		err = fmt.Errorf("no config for: %s")
		return
	}

	if cfg.RefreshToken == "" || cfg.SSOURL == "" || cfg.Username == "" {
		return cfg, fmt.Errorf("incomplete config")
	}

	return configs[environ], nil
}
