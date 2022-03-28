package config

import (
	"fmt"
	"time"

	"github.com/halimath/raspidoor/daemon/internal/gatekeeper"
	"github.com/halimath/raspidoor/daemon/internal/gpio"
	"github.com/halimath/raspidoor/daemon/internal/sip"
	"github.com/halimath/raspidoor/systemd/logging"

	"github.com/halimath/appconf"
)

type (
	SIPServer struct {
		// The SIP host (your phone router)
		Host string

		// The SIP port (should be 5060 by default)
		Port int

		// The SIP user name to authenticate with
		User string

		// The SIP password to authenticate with
		Password string

		// Whether to dump protocol logs
		Debug bool
	}

	SIP struct {
		// The caller's SIP address
		Caller string

		// The callee's SIP address
		Callee string

		// Max duration to ring the users phone
		MaxRingingTime time.Duration

		// Server settings
		Server SIPServer
	}

	// StatusLED defines the config for the status led.
	StatusLED struct {
		// GPIO number (not the physical pin) to connect the status LED on
		GPIO int

		// Duration the LED should blink when a bell push is pressed
		BlinkDuration time.Duration
	}

	// ExternalBell is anything that can be switched on and off.
	ExternalBell struct {
		// GPIO number (not the physical pin) to connect the status LED on
		GPIO int

		// Duration to ring the external bell (keep the relay open) when a bell push is pressed
		RingDuration time.Duration
	}

	// BellPush defines the individual bell pushes the system should react on.
	BellPush struct {
		// A human readable label for the bell push
		Label string

		// GPIO number (not the physical pin) to connect the bell push IN to
		GPIO int
	}

	// Controller defines the config for the controller.
	Controller struct {
		// Socket defines the path of the Unix socket to receive commands on.
		Socket string
	}

	// Logging defines the log configuration
	Logging struct {
		// Target defines the log target; must be either STDOUT or SYSLOG
		Target string

		// Debug defines, whether to output debug messages
		Debug bool
	}

	// Config is the root of the config settings.
	Config struct {
		SIP          SIP
		StatusLED    StatusLED
		ExternalBell ExternalBell
		BellPushes   []BellPush
		Logging      Logging
		Controller   Controller
		DisableGPIO  bool
	}
)

func (c Config) NewLogger() (logging.Logger, error) {
	if c.Logging.Target == "syslog" {
		return logging.Syslog("raspidoord")
	}
	return logging.Stdout(), nil
}

func (c Config) GatekeeperOptions() (gatekeeper.Options, error) {
	var err error

	caller, err := sip.ParseURI(c.SIP.Caller)
	if err != nil {
		return gatekeeper.Options{}, err
	}

	callee, err := sip.ParseURI(c.SIP.Callee)
	if err != nil {
		return gatekeeper.Options{}, err
	}

	transport := &sip.TCPTransport{
		DumpRoundTrips: c.SIP.Server.Debug,
	}

	bellPushes := make([]gatekeeper.BellPushOptions, len(c.BellPushes))
	for i, p := range c.BellPushes {
		var input gpio.DigitalInput
		if c.DisableGPIO {
			input = gpio.NewNOOPDigitalInput()
		} else {
			input, err = gpio.NewPushButton(gpio.DefaultChip, p.GPIO, gpio.TypePullUp)
			if err != nil {
				return gatekeeper.Options{}, err
			}
		}

		bellPushes[i] = gatekeeper.BellPushOptions{
			Label: p.Label,
			Input: input,
		}
	}

	var led gpio.DigitalOutput
	if c.DisableGPIO {
		led = gpio.NewNOOPDigitalOutput()
	} else {
		led, err = gpio.NewDigitalOutput(gpio.DefaultChip, c.StatusLED.GPIO)
		if err != nil {
			return gatekeeper.Options{}, err
		}
	}

	var externalBell gpio.DigitalOutput
	if c.DisableGPIO {
		externalBell = gpio.NewNOOPDigitalOutput()
	} else {
		externalBell, err = gpio.NewDigitalOutput(gpio.DefaultChip, c.ExternalBell.GPIO)
		if err != nil {
			return gatekeeper.Options{}, err
		}
	}

	return gatekeeper.Options{
		StatusLED:   led,
		LEDDuration: c.StatusLED.BlinkDuration,
		BellPushes:  bellPushes,
		Bells: []gatekeeper.BellOptions{
			gatekeeper.NewExternalBell("External Bell", externalBell, c.ExternalBell.RingDuration),
			gatekeeper.NewPhoneBell("SIP Phone", caller, callee, c.SIP.MaxRingingTime, transport, []sip.AuthenticationHandler{sip.NewDigestHandler(c.SIP.Server.User, c.SIP.Server.Password)}),
		},
	}, nil
}

func ReadConfig() (*Config, error) {
	return readConfigFromFiles(
		"/etc/raspidoor/raspidoord.yaml",
		"/etc/raspidoor/conf.d/raspidoord.yaml",
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
