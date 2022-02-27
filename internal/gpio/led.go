package gpio

import (
	"errors"
	"sync"
	"time"

	"github.com/warthog618/gpiod"
)

type OnOffOutput struct {
	line  *gpiod.Line
	lock  sync.RWMutex
	state bool
}

func NewOnOffOutput(chip string, gpioNumber int) (*OnOffOutput, error) {
	line, err := gpiod.RequestLine(chip, gpioNumber, gpiod.AsOutput(0))
	if err != nil {
		return nil, err
	}

	return &OnOffOutput{
		line:  line,
		state: false,
	}, nil
}

func (o *OnOffOutput) State() bool {
	o.lock.RLock()
	defer o.lock.RUnlock()

	return o.state
}

func (o *OnOffOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	return o.line.Close()
}

func (o *OnOffOutput) On() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.state = true
	return o.line.SetValue(1)
}

func (o *OnOffOutput) OnFor(dur time.Duration) error {
	timer := time.NewTimer(dur)

	go func() {
		defer timer.Stop()
		<-timer.C
		o.Off()
	}()

	return o.On()
}

func (o *OnOffOutput) Off() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.state = false
	return o.line.SetValue(0)
}

func (o *OnOffOutput) Toggle() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.state {
		o.state = false
		return o.line.SetValue(0)
	}

	o.state = true
	return o.line.SetValue(1)
}

var ErrAlreadyBlinking = errors.New("already blinking")

type LED struct {
	*OnOffOutput
	lock sync.Mutex
	line *gpiod.Line
}

func NewLED(chip string, gpioNumber int) (*LED, error) {
	output, err := NewOnOffOutput(chip, gpioNumber)
	if err != nil {
		return nil, err
	}
	return &LED{
		OnOffOutput: output,
	}, nil
}

func (l *LED) Blink(duration time.Duration) (chan<- struct{}, error) {
	if ok := l.lock.TryLock(); !ok {
		return nil, ErrAlreadyBlinking
	}
	ticker := time.NewTicker(duration)

	stopChan := make(chan struct{})

	s := l.state

	go func() {
		defer ticker.Stop()
		defer l.lock.Unlock()

		for {
			select {
			case <-stopChan:
				if s {
					l.On()
				} else {
					l.Off()
				}
				return

			case <-ticker.C:
				l.Toggle()
			}
		}
	}()

	return stopChan, nil
}

func (l *LED) BlinkFor(blinkDuration, onDuration time.Duration) error {
	stopChan, err := l.Blink(onDuration)
	if err != nil {
		return err
	}

	go func() {
		timer := time.NewTimer(blinkDuration)
		<-timer.C
		stopChan <- struct{}{}
	}()

	return nil
}
