package config

import (
	"fmt"

	"github.com/halimath/appconf"
	"github.com/halimath/raspidoor/systemd/logging"
)

type (
	// Logging defines the log configuration
	Logging struct {
		// Target defines the log target; must be either STDOUT or SYSLOG
		Target string

		// Debug defines, whether to output debug messages
		Debug bool
	}

	Config struct {
		Address string
		Socket  string
		Logging Logging
	}
)

func (c Config) NewLogger() (logging.Logger, error) {
	if c.Logging.Target == "syslog" {
		return logging.Syslog("raspidoorwebapp")
	}
	return logging.Stdout(), nil
}

func ReadConfig() (*Config, error) {
	return readConfigFromFiles(
		"/etc/raspidoor/raspidoorwebd.yaml",
		"/etc/raspidoor/conf.d/raspidoorwebd.yaml",
	)
}

func ReadConfigFromFile(n string) (*Config, error) {
	return readConfigFromFiles(n)
}

func readConfigFromFiles(names ...string) (*Config, error) {
	loaders := make([]appconf.Loader, len(names))
	for i, n := range names {
		loaders[i] = appconf.YAMLFile(n, i == 0)
	}

	ac, err := appconf.New(loaders...)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %s", err)
	}

	var c Config
	err = ac.Bind(&c)
	return &c, err
}
