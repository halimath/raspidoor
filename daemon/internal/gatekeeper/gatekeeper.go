package gatekeeper

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/halimath/raspidoor/daemon/internal/gpio"
	"github.com/halimath/raspidoor/systemd/logging"
)

const (
	ItemInfoLabelExternalBell = "external bell"
	ItemInfoLabelPhone        = "phone"
)

var (
	ErrNotFound = errors.New("not found")
)

type (
	BellPushOptions struct {
		Label string
		Input gpio.DigitalInput
	}

	BellOptions struct {
		Label  string
		Ringer Ringer
	}

	Options struct {
		StatusLED   gpio.DigitalOutput
		LEDDuration time.Duration

		BellPushes []BellPushOptions
		Bells      []BellOptions
	}

	bellPush struct {
		enabled bool
		label   string
		btn     gpio.DigitalInput
	}

	bell struct {
		enabled bool
		label   string
		ringer  Ringer
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

		statusLED  *gpio.LED
		bells      []*bell
		bellPushes []*bellPush

		logger logging.Logger

		lock sync.RWMutex
	}
)

func (b *bell) Ring(logger logging.Logger) {
	b.ringer.Ring(logger)
}

func (b *bell) Close() error {
	return b.ringer.Close()
}

func New(opts Options, logger logging.Logger) (*Gatekeeper, error) {
	g := &Gatekeeper{
		opts: opts,

		statusLED:  gpio.NewLED(opts.StatusLED),
		bells:      make([]*bell, len(opts.Bells)),
		bellPushes: make([]*bellPush, len(opts.BellPushes)),
		logger:     logger,
	}

	for i, p := range opts.BellPushes {
		g.bellPushes[i] = &bellPush{
			enabled: true,
			label:   p.Label,
			btn:     p.Input,
		}
		func(i int) {
			g.bellPushes[i].btn.AddCallback(func(pressed bool) {
				if !pressed {
					return
				}

				g.bellPushPressed(i)
			})
		}(i)
	}

	for i, b := range opts.Bells {
		g.bells[i] = &bell{
			enabled: true,
			label:   b.Label,
			ringer:  b.Ringer,
		}
	}

	return g, nil
}

func (g *Gatekeeper) Start() {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.logger.Info("Starting gatekeeper")
	g.statusLED.On()
}

func (g *Gatekeeper) Close() error {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.logger.Info("Shutting down gatekeeper")

	g.statusLED.Off()
	if err := g.statusLED.Close(); err != nil {
		return err
	}

	for _, p := range g.bellPushes {
		if err := p.btn.Close(); err != nil {
			return err
		}
	}

	for _, b := range g.bells {
		if err := b.Close(); err != nil {
			return err
		}
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

	if err := g.statusLED.BlinkFor(g.opts.LEDDuration, 100*time.Millisecond); err != nil {
		g.logger.Error("failed to blink status led: %s", err)
	}

	for _, b := range g.bells {
		if b.enabled {
			b.Ring(g.logger)
		}
	}
}

func (g *Gatekeeper) SetBellPushState(index int, enabled bool) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if index >= len(g.bellPushes) {
		return fmt.Errorf("%w: bell push %d", ErrNotFound, index)
	}

	g.bellPushes[index].enabled = enabled

	return nil
}

func (g *Gatekeeper) SetBellState(index int, enabled bool) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if index >= len(g.bellPushes) {
		return fmt.Errorf("%w: bell %d", ErrNotFound, index)
	}

	g.bells[index].enabled = enabled
	return nil
}

func (g *Gatekeeper) Info() Info {
	g.lock.RLock()
	defer g.lock.RUnlock()

	i := Info{
		Bells:      make([]ItemInfo, len(g.bells)),
		BellPushes: make([]ItemInfo, len(g.bellPushes)),
	}

	for idx, p := range g.bellPushes {
		i.BellPushes[idx] = ItemInfo{
			Label:   p.label,
			Enabled: p.enabled,
		}
	}

	for idx, b := range g.bells {
		i.Bells[idx] = ItemInfo{
			Label:   b.label,
			Enabled: b.enabled,
		}
	}

	return i
}
