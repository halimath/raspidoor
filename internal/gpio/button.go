package gpio

import (
	"time"

	"github.com/warthog618/gpiod"
)

// PullType defines the type the switch is physically pulled to.
type PullType int

const (
	// TypePullUp defines the switch as pull-up wiring; the input pin is pulled to "high" and falls to "GND"
	// when the button is pushed.
	TypePullUp PullType = iota

	// TypePullDown defines the switch as pull-down wiring; the input pin is pulled to "GND" and rises to
	// "HIGH" when the button is pushed.
	TypePullDown
)

type Button struct {
	line      *gpiod.Line
	pullType  PullType
	callbacks []ButtonCallback
}

type ButtonCallback func(pressed bool)

func NewButton(chip string, gpioNumber int, pullType PullType) (*Button, error) {
	b := &Button{
		pullType:  pullType,
		callbacks: make([]ButtonCallback, 0, 5),
	}

	line, err := gpiod.RequestLine(chip, gpioNumber, gpiod.WithEventHandler(b.handleEvent), gpiod.WithDebounce(200*time.Millisecond), gpiod.WithBothEdges)
	if err != nil {
		return nil, err
	}

	b.line = line

	return b, nil
}

func (b *Button) handleEvent(evt gpiod.LineEvent) {
	var pressed bool

	if b.pullType == TypePullUp {
		pressed = evt.Type == gpiod.LineEventFallingEdge
	} else {
		pressed = evt.Type == gpiod.LineEventRisingEdge
	}

	for _, cb := range b.callbacks {
		cb(pressed)
	}
}

func (b *Button) AddCallback(cb ButtonCallback) {
	b.callbacks = append(b.callbacks, cb)
}

func (b *Button) Close() error {
	return b.line.Close()
}
