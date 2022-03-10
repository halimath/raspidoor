package gatekeeper

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/halimath/raspidoor/internal/gpio"
	"github.com/halimath/raspidoor/internal/sip"
	"github.com/halimath/raspidoor/logging"
)

const (
	ItemInfoLabelExternalBell = "external bell"
	ItemInfoLabelPhone        = "phone"
)

var (
	ErrNoSuchBellPush = errors.New("no such bell push")
)

type (
	SIPOptions struct {
		Caller       sip.URI
		Callee       sip.URI
		RoundTripper sip.RoundTripper
		AuthHandler  []sip.AuthenticationHandler
	}

	GPIOOptions struct {
		Chip string
		GPIO int
	}

	GPIOOutputOptions struct {
		GPIOOptions
		Duration time.Duration
	}

	BellPushOptions struct {
		GPIOOptions
		Label string
	}

	Options struct {
		SIP          SIPOptions
		StatusLED    GPIOOutputOptions
		ExternalBell GPIOOutputOptions
		BellPushes   []BellPushOptions
		DisableGPIO  bool
	}

	bellPush struct {
		enabled bool
		label   string
		btn     *gpio.Button
	}

	externalBell struct {
		enabled bool
		dur     time.Duration
		out     *gpio.OnOffOutput
		logger  logging.Logger
	}

	phoneBell struct {
		enabled      bool
		caller       sip.URI
		callee       sip.URI
		roundTripper sip.RoundTripper
		authHandler  []sip.AuthenticationHandler
		logger       logging.Logger
	}

	ItemInfo struct {
		Label   string
		Enabled bool
	}

	Info struct {
		BellPushes []ItemInfo
		Bells      []ItemInfo
	}

	Gatekeeper struct {
		opts Options

		statusLED    *gpio.LED
		externalBell externalBell
		phoneBell    *phoneBell
		bellPushes   []bellPush

		logger logging.Logger

		lock sync.RWMutex
	}
)

func (e *externalBell) ring() {
	if err := e.out.OnFor(e.dur); err != nil {
		e.logger.Error("failed to ring external bell: %s", err)
	}
}

func (p *phoneBell) Close() error {
	return p.roundTripper.Close()
}

func (p *phoneBell) ring() {
	go func() {
		d := sip.NewDialog(p.roundTripper, p.caller, p.authHandler...)
		if err := d.Ring(p.callee); err != nil {
			p.logger.Error("failed to ring SIP phone: %s", err)
		}
	}()
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

		statusLED: led,

		phoneBell: &phoneBell{
			enabled:      true,
			caller:       opts.SIP.Caller,
			callee:       opts.SIP.Callee,
			roundTripper: opts.SIP.RoundTripper,
			authHandler:  opts.SIP.AuthHandler,
			logger:       logger,
		},

		externalBell: externalBell{
			enabled: true,
			dur:     opts.ExternalBell.Duration,
			out:     extlBell,
			logger:  logger,
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
	g.lock.Lock()
	defer g.lock.Unlock()

	g.logger.Info("Starting gatekeeper")
	if !g.opts.DisableGPIO {
		g.statusLED.On()
	}
}

func (g *Gatekeeper) Close() error {
	g.lock.Lock()
	defer g.lock.Unlock()

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

	if err := g.phoneBell.Close(); err != nil {
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
	g.lock.RLock()
	defer g.lock.RUnlock()

	if !g.opts.DisableGPIO {
		if err := g.statusLED.BlinkFor(g.opts.StatusLED.Duration, 100*time.Millisecond); err != nil {
			g.logger.Error("failed to blink status led: %s", err)
		}

		if g.externalBell.enabled {
			g.externalBell.ring()
		}
	}

	if g.phoneBell.enabled {
		g.phoneBell.ring()
	}
}

func (g *Gatekeeper) SetBellPushState(index int, enabled bool) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if index >= len(g.bellPushes) {
		return fmt.Errorf("%w: %d", ErrNoSuchBellPush, index)
	}

	g.bellPushes[index].enabled = enabled

	return nil
}

func (g *Gatekeeper) SetExternalBellState(enabled bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.externalBell.enabled = enabled
}

func (g *Gatekeeper) SetPhoneBellState(enabled bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.phoneBell.enabled = enabled
}

func (g *Gatekeeper) Info() Info {
	g.lock.RLock()
	defer g.lock.RUnlock()

	i := Info{
		Bells: []ItemInfo{
			{
				Label:   ItemInfoLabelExternalBell,
				Enabled: g.externalBell.enabled,
			},
			{
				Label:   ItemInfoLabelPhone,
				Enabled: g.phoneBell.enabled,
			},
		},
		BellPushes: make([]ItemInfo, len(g.bellPushes)),
	}

	for idx, p := range g.bellPushes {
		i.BellPushes[idx] = ItemInfo{
			Label:   p.label,
			Enabled: p.enabled,
		}
	}

	return i
}

func chip(o GPIOOptions) string {
	if len(o.Chip) == 0 {
		return gpio.DefaultChip
	}

	return o.Chip
}
