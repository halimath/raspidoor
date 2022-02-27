package gatekeeper

import (
	"errors"
	"fmt"
	"time"

	"github.com/halimath/raspidoor/internal/gpio"
	"github.com/halimath/raspidoor/internal/sip"
	"github.com/halimath/raspidoor/logging"
)

var (
	ErrNoSuchBellPush = errors.New("no such bell push")
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
	DisableGPIO  bool
}

type bellPush struct {
	enabled bool
	label   string
	btn     *gpio.Button
}

type externalBell struct {
	enabled bool
	out     *gpio.OnOffOutput
}

type Gatekeeper struct {
	opts         Options
	caller       sip.URI
	callee       sip.URI
	roundTripper sip.RoundTripper
	authHandler  []sip.AuthenticationHandler

	statusLED    *gpio.LED
	externalBell externalBell
	bellPushes   []bellPush

	logger logging.Logger
}

func New(opts Options, logger logging.Logger) (*Gatekeeper, error) {
	var led *gpio.LED
	var extlBell *gpio.OnOffOutput
	var err error

	if !opts.DisableGPIO {
		led, err = gpio.NewLED(chip(opts.StatusLED.GPIOOptions), opts.StatusLED.GPIO)
		if err != nil {
			return nil, fmt.Errorf("failed to create status led: %w", err)
		}

		extlBell, err = gpio.NewOnOffOutput(chip(opts.ExternalBell.GPIOOptions), opts.ExternalBell.GPIO)
		if err != nil {
			return nil, fmt.Errorf("failed to create external bell: %w", err)
		}
	}

	g := &Gatekeeper{
		opts: opts,

		caller:       opts.SIP.Caller,
		callee:       opts.SIP.Callee,
		roundTripper: opts.SIP.RoundTripper,
		authHandler:  opts.SIP.AuthHandler,
		statusLED:    led,
		externalBell: externalBell{
			enabled: true,
			out:     extlBell,
		},

		logger: logger,
	}

	pushes := make([]bellPush, len(opts.BellPushes))
	for i, p := range opts.BellPushes {
		var sw *gpio.Button
		if !opts.DisableGPIO {
			sw, err = func(i int, p BellPushOptions) (*gpio.Button, error) {
				btn, err := gpio.NewButton(chip(p.GPIOOptions), p.GPIO, gpio.TypePullUp)
				if err != nil {
					return nil, err
				}

				btn.AddCallback(func(pressed bool) {
					if !pressed {
						return
					}

					g.bellPushPressed(i)
				})

				return btn, nil
			}(i, p)
			if err != nil {
				return nil, fmt.Errorf("failed to create bell switch '%s': %w", p.Label, err)
			}
		}

		pushes[i] = bellPush{
			enabled: true,
			label:   p.Label,
			btn:     sw,
		}
	}

	g.bellPushes = pushes

	return g, nil
}

func (g *Gatekeeper) Start() {
	g.logger.Info("Starting gatekeeper")
	if !g.opts.DisableGPIO {
		g.statusLED.On()
	}
}

func (g *Gatekeeper) Close() error {
	g.logger.Info("Shutting down gatekeeper")

	if !g.opts.DisableGPIO {
		g.statusLED.Off()
		if err := g.statusLED.Close(); err != nil {
			return err
		}

		for _, p := range g.bellPushes {
			if err := p.btn.Close(); err != nil {
				return err
			}
		}
	}

	if err := g.roundTripper.Close(); err != nil {
		return err
	}

	return g.logger.Close()
}

func (g *Gatekeeper) bellPushPressed(idx int) {
	g.logger.Info("Pressed bell push %d: %s", idx, g.bellPushes[idx].label)

	if !g.bellPushes[idx].enabled {
		g.logger.Info("Bell push %d disabled; not ringing", idx)
		return
	}

	g.Ring()
}

func (g *Gatekeeper) Ring() {
	if !g.opts.DisableGPIO {
		if err := g.statusLED.BlinkFor(g.opts.StatusLED.Duration, 100*time.Millisecond); err != nil {
			g.logger.Error("failed to blink status led: %s", err)
		}

		if g.externalBell.enabled {
			if err := g.externalBell.out.OnFor(g.opts.ExternalBell.Duration); err != nil {
				g.logger.Error("failed to activate external bell: %s", err)
			}
		}
	}

	go func() {
		d := sip.NewDialog(g.roundTripper, g.caller, g.authHandler...)
		if err := d.Ring(g.callee); err != nil {
			g.logger.Error("failed to ring SIP phone: %s", err)
		}
	}()
}

func (g *Gatekeeper) SetBellPushState(index int, enabled bool) error {
	if index >= len(g.bellPushes) {
		return fmt.Errorf("%w: %d", ErrNoSuchBellPush, index)
	}

	g.bellPushes[index].enabled = enabled

	return nil
}

func chip(o GPIOOptions) string {
	if len(o.Chip) == 0 {
		return gpio.DefaultChip
	}

	return o.Chip
}
