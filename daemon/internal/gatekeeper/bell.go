package gatekeeper

import (
	"time"

	"github.com/halimath/raspidoor/daemon/internal/gpio"
	"github.com/halimath/raspidoor/daemon/internal/sip"
	"github.com/halimath/raspidoor/systemd/logging"
)

type (
	Ringer interface {
		Ring(logger logging.Logger)
		Close() error
	}

	externalBell struct {
		dur time.Duration
		out gpio.DigitalOutput
	}

	phoneBell struct {
		caller       sip.URI
		callee       sip.URI
		roundTripper sip.RoundTripper
		authHandler  []sip.AuthenticationHandler
	}
)

func (e *externalBell) Ring(logger logging.Logger) {
	if err := gpio.OnFor(e.out, e.dur); err != nil {
		logger.Error("failed to ring external bell: %s", err)
	}
}

func (e *externalBell) Close() error { return e.out.Close() }

func (p *phoneBell) Ring(logger logging.Logger) {
	go func() {
		d := sip.NewDialog(p.roundTripper, p.caller, p.authHandler...)
		if err := d.Ring(p.callee); err != nil {
			logger.Error("failed to ring SIP phone: %s", err)
		}
	}()
}

func (p *phoneBell) Close() error {
	return p.roundTripper.Close()
}

func NewExternalBell(label string, out gpio.DigitalOutput, dur time.Duration) BellOptions {
	return BellOptions{
		Label: label,
		Ringer: &externalBell{
			dur: dur,
			out: out,
		},
	}
}

func NewPhoneBell(label string,
	caller sip.URI,
	callee sip.URI,
	roundTripper sip.RoundTripper,
	authHandler []sip.AuthenticationHandler,
) BellOptions {
	return BellOptions{
		Label: label,
		Ringer: &phoneBell{
			caller:       caller,
			callee:       callee,
			roundTripper: roundTripper,
			authHandler:  authHandler,
		},
	}
}
