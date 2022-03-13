package config

import (
	"fmt"
	"io"
	"os"

	"github.com/halimath/raspidoor/systemd/logging"
	"gopkg.in/yaml.v2"
)

type (
	// Logging defines the log configuration
	Logging struct {
		// Target defines the log target; must be either STDOUT or SYSLOG
		Target string `yaml:"target"`

		// Debug defines, whether to output debug messages
		Debug bool `yaml: "debug"`
	}

	Config struct {
		Address string  `yaml:"address"`
		Socket  string  `yaml:"socket"`
		Logging Logging `yaml:"logging"`
	}
)

func (c Config) NewLogger() (logging.Logger, error) {
	if c.Logging.Target == "syslog" {
		return logging.Syslog("raspidoorwebapp")
	}
	return logging.Stdout(), nil
}

func ReadConfigFile(n string) (*Config, error) {
	file, err := os.Open(n)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	defer file.Close()

	return ReadConfig(file)
}

func ReadConfig(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
