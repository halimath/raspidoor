package gpio

import (
	"errors"
	"sync"
	"time"

	"github.com/warthog618/gpiod"
)

type digitalOutput struct {
	line  *gpiod.Line
	lock  sync.RWMutex
	state bool
}

func NewDigitalOutput(chip string, gpioNumber int) (DigitalOutput, error) {
	line, err := gpiod.RequestLine(chip, gpioNumber, gpiod.AsOutput(0))
	if err != nil {
		return nil, err
	}

	return &digitalOutput{
		line:  line,
		state: false,
	}, nil
}

func (d *digitalOutput) State() bool {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.state
}

func (d *digitalOutput) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.line.Close()
}

func (d *digitalOutput) On() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.state = true
	return d.line.SetValue(1)
}

func (d *digitalOutput) Off() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.state = false
	return d.line.SetValue(0)
}

var ErrAlreadyBlinking = errors.New("already blinking")

type LED struct {
	DigitalOutput
	lock sync.Mutex
	line *gpiod.Line
}

func NewLED(o DigitalOutput) *LED {
	return &LED{
		DigitalOutput: o,
	}
}

func (l *LED) Blink(duration time.Duration) (chan<- struct{}, error) {
	if ok := l.lock.TryLock(); !ok {
		return nil, ErrAlreadyBlinking
	}
	ticker := time.NewTicker(duration)

	stopChan := make(chan struct{})

	s := l.State()

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
				Toggle(l)
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
