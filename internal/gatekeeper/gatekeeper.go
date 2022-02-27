package gatekeeper

import (
	"fmt"
	"time"

	"github.com/halimath/raspidoor/internal/gpio"
	"github.com/halimath/raspidoor/internal/logging"
	"github.com/halimath/raspidoor/internal/sip"
)

type SIPOptions struct {
	Caller       sip.URI
	Callee       sip.URI
	RoundTripper sip.RoundTripper
	AuthHandler  []sip.AuthenticationHandler
}

type GPIOOptions struct {
	Chip string
	GPIO int
}

type GPIOOutputOptions struct {
	GPIOOptions
	Duration time.Duration
}

type BellPushOptions struct {
	GPIOOptions
	Label string
}

type Options struct {
	SIP          SIPOptions
	StatusLED    GPIOOutputOptions
	ExternalBell GPIOOutputOptions
	BellPushes   []BellPushOptions
	Logger       logging.Logger
}

type Gatekeeper struct {
	opts         Options
	caller       sip.URI
	callee       sip.URI
	roundTripper sip.RoundTripper
	authHandler  []sip.AuthenticationHandler

	statusLED    *gpio.LED
	externalBell *gpio.OnOffOutput
	bellPushes   []*gpio.Button
}

func New(opts Options) (*Gatekeeper, error) {
	led, err := gpio.NewLED(chip(opts.StatusLED.GPIOOptions), opts.StatusLED.GPIO)
	if err != nil {
		return nil, fmt.Errorf("failed to create status led: %w", err)
	}

	externalBell, err := gpio.NewOnOffOutput(chip(opts.ExternalBell.GPIOOptions), opts.ExternalBell.GPIO)
	if err != nil {
		return nil, fmt.Errorf("failed to create external bell: %w", err)
	}

	g := &Gatekeeper{
		opts: opts,

		caller:       opts.SIP.Caller,
		callee:       opts.SIP.Callee,
		roundTripper: opts.SIP.RoundTripper,
		authHandler:  opts.SIP.AuthHandler,
		statusLED:    led,
		externalBell: externalBell,
	}

	pushes := make([]*gpio.Button, len(opts.BellPushes))
	for i, p := range opts.BellPushes {
		sw, err := func(i int, p BellPushOptions) (*gpio.Button, error) {
			btn, err := gpio.NewButton(chip(p.GPIOOptions), p.GPIO, gpio.TypePullUp)
			if err != nil {
				return nil, err
			}

			btn.AddCallback(func(pressed bool) {
				if !pressed {
					return
				}

				g.bellPushPressed(i, p.Label)
			})

			return btn, nil
		}(i, p)
		if err != nil {
			return nil, fmt.Errorf("failed to create bell switch '%s': %w", p.Label, err)
		}

		pushes[i] = sw
	}

	g.bellPushes = pushes

	return g, nil
}

func (g *Gatekeeper) Start() {
	g.opts.Logger.Info("Starting gatekeeper")
	g.statusLED.On()
}

func (g *Gatekeeper) Close() error {
	g.opts.Logger.Info("Shutting down gatekeeper")

	g.statusLED.Off()
	if err := g.statusLED.Close(); err != nil {
		return err
	}

	for _, p := range g.bellPushes {
		if err := p.Close(); err != nil {
			return err
		}
	}

	if err := g.roundTripper.Close(); err != nil {
		return err
	}

	return g.opts.Logger.Close()
}

func (g *Gatekeeper) bellPushPressed(idx int, label string) {
	g.opts.Logger.Info("Pressed bell push %s (%d)\n", label, idx)

	g.Ring()
}

func (g *Gatekeeper) Ring() {
	if err := g.statusLED.BlinkFor(g.opts.StatusLED.Duration, 100*time.Millisecond); err != nil {
		g.opts.Logger.Error("failed to blink status led: %s", err)
	}

	if err := g.externalBell.OnFor(g.opts.ExternalBell.Duration); err != nil {
		g.opts.Logger.Error("failed to activate external bell: %s", err)
	}

	go func() {
		d := sip.NewDialog(g.roundTripper, g.caller, g.authHandler...)
		if err := d.Ring(g.callee); err != nil {
			g.opts.Logger.Error("failed to ring SIP phone: %s", err)
		}
	}()
}

func chip(o GPIOOptions) string {
	if len(o.Chip) == 0 {
		return gpio.DefaultChip
	}

	return o.Chip
}
