package config

import (
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestReadConfig(t *testing.T) {
	r := strings.NewReader(`
sip:
  caller: "sip:caller@registrar.example.com"
  callee: "sip:callee@registrar.example.com"
  server:
    host: "registrar.example.com"
    port: 5060
    user: caller
    password: "password001"
    debug: False
statusLed:
  gpio: 23
  blinkDuration: 2s
externalBell:
  gpio: 25
  ringDuration: 2s
bellPushes:
- label: Main door
  gpio: 24
logging:
  debug: True	  
`)

	config, err := ReadConfig(r)
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(*config, Config{
		SIP: SIP{
			Caller: "sip:caller@registrar.example.com",
			Callee: "sip:callee@registrar.example.com",
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
