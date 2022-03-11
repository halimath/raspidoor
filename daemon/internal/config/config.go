package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/halimath/raspidoor/daemon/internal/gatekeeper"
	"github.com/halimath/raspidoor/daemon/internal/gpio"
	"github.com/halimath/raspidoor/daemon/internal/logging"
	"github.com/halimath/raspidoor/daemon/internal/sip"
	"gopkg.in/yaml.v2"
)

type (
	SIPServer struct {
		// The SIP host (your phone router)
		Host string `yaml:"host"`

		// The SIP port (should be 5060 by default)
		Port int `yaml:"port"`

		// The SIP user name to authenticate with
		User string `yaml:"user"`

		// The SIP password to authenticate with
		Password string `yaml:"password"`

		// Whether to dump protocol logs
		Debug bool `yaml:"debug"`
	}

	SIP struct {
		// The caller's SIP address
		Caller string `yaml:"caller"`

		// The callee's SIP address
		Callee string `yaml:"callee"`

		// Server settings
		Server SIPServer `yaml:"server"`
	}

	// StatusLED defines the config for the status led.
	StatusLED struct {
		// GPIO number (not the physical pin) to connect the status LED on
		GPIO int `yaml:"gpio"`

		// Duration the LED should blink when a bell push is pressed
		BlinkDuration time.Duration `yaml:"blinkDuration"`
	}

	// ExternalBell is anything that can be switched on and off.
	ExternalBell struct {
		// GPIO number (not the physical pin) to connect the status LED on
		GPIO int `yaml:"gpio"`

		// Duration to ring the external bell (keep the relay open) when a bell push is pressed
		RingDuration time.Duration `yaml:"ringDuration"`
	}

	// BellPush defines the individual bell pushes the system should react on.
	BellPush struct {
		// A human readable label for the bell push
		Label string `yaml:"label"`

		// GPIO number (not the physical pin) to connect the bell push IN to
		GPIO int `yaml:"gpio"`
	}

	// Controller defines the config for the controller.
	Controller struct {
		// Socket defines the path of the Unix socket to receive commands on.
		Socket string `yaml:"socket"`
	}

	// Logging defines the log configuration
	Logging struct {
		// Target defines the log target; must be either STDOUT or SYSLOG
		Target string `yaml:"target"`

		// Debug defines, whether to output debug messages
		Debug bool `yaml: "debug"`
	}

	// Config is the root of the config settings.
	Config struct {
		SIP          SIP          `yaml:"sip"`
		StatusLED    StatusLED    `yaml:"statusLed"`
		ExternalBell ExternalBell `yaml:"externalBell"`
		BellPushes   []BellPush   `yaml:"bellPushes"`
		Logging      Logging      `yaml:"logging"`
		Controller   Controller   `yaml:"controller"`
		DisableGPIO  bool         `yaml:"disableGpio"`
	}
)

func (c Config) NewLogger() (logging.Logger, error) {
	if c.Logging.Target == "syslog" {
		return logging.Syslog()
	}
	return logging.Stdout(), nil
}

func (c Config) GatekeeperOptions() (gatekeeper.Options, error) {
	caller, err := sip.ParseURI(c.SIP.Caller)
	if err != nil {
		return gatekeeper.Options{}, err
	}

	callee, err := sip.ParseURI(c.SIP.Callee)
	if err != nil {
		return gatekeeper.Options{}, err
	}

	bellPushes := make([]gatekeeper.BellPushOptions, len(c.BellPushes))
	for i, p := range c.BellPushes {
		bellPushes[i] = gatekeeper.BellPushOptions{
			GPIOOptions: gatekeeper.GPIOOptions{
				Chip: gpio.DefaultChip,
				GPIO: p.GPIO,
			},
			Label: p.Label,
		}
	}

	transport := sip.NewTCPTransport()
	if c.SIP.Server.Debug {
		transport.DumpRoundTrips = true
	}

	return gatekeeper.Options{
		SIP: gatekeeper.SIPOptions{
			Caller:       caller,
			Callee:       callee,
			RoundTripper: transport,
			AuthHandler:  []sip.AuthenticationHandler{sip.NewDigestHandler(c.SIP.Server.User, c.SIP.Server.Password)},
		},
		StatusLED: gatekeeper.GPIOOutputOptions{
			GPIOOptions: gatekeeper.GPIOOptions{
				Chip: gpio.DefaultChip,
				GPIO: c.StatusLED.GPIO,
			},
			Duration: c.StatusLED.BlinkDuration,
		},
		ExternalBell: gatekeeper.GPIOOutputOptions{
			GPIOOptions: gatekeeper.GPIOOptions{
				Chip: gpio.DefaultChip,
				GPIO: c.ExternalBell.GPIO,
			},
			Duration: c.ExternalBell.RingDuration,
		},
		BellPushes:  bellPushes,
		DisableGPIO: c.DisableGPIO,
	}, nil
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
