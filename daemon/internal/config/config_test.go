package config

import (
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestReadConfig(t *testing.T) {
	config, err := ReadConfigFromFile("./testdata/config.yaml")

	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(*config, Config{
		SIP: SIP{
			Caller:         "sip:caller@registrar.example.com",
			Callee:         "sip:callee@registrar.example.com",
			MaxRingingTime: 5 * time.Second,
			Server: SIPServer{
				Host:     "registrar.example.com",
				Port:     5060,
				User:     "caller",
				Password: "password001",
				Debug:    false,
			},
		},
		StatusLED: StatusLED{
			GPIO:          23,
			BlinkDuration: 2 * time.Second,
		},
		ExternalBell: ExternalBell{
			GPIO:         25,
			RingDuration: 2 * time.Second,
		},
		BellPushes: []BellPush{
			{
				Label: "Main door",
				GPIO:  24,
			},
		},
		Logging: Logging{
			Debug: true,
		},
	}); diff != nil {
		t.Error(diff)
	}
}
