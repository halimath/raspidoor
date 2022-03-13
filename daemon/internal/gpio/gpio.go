package gpio

import "time"

const DefaultChip = "gpiochip0"

type (
	DigitalOutput interface {
		State() bool
		Close() error
		On() error
		Off() error
	}

	DigitalInputCallback func(state bool)

	DigitalInput interface {
		Close() error
		AddCallback(cb DigitalInputCallback)
	}

	dummyInput struct{}

	dummyOutput struct {
		s bool
	}
)

func Toggle(out DigitalOutput) error {
	if out.State() {
		return out.Off()
	}
	return out.On()
}

func OnFor(out DigitalOutput, dur time.Duration) error {
	timer := time.NewTimer(dur)

	go func() {
		defer timer.Stop()
		<-timer.C
		out.Off()
	}()

	return out.On()
}

func (d *dummyOutput) State() bool  { return d.s }
func (d *dummyOutput) Close() error { return nil }
func (d *dummyOutput) On() error {
	d.s = true
	return nil
}
func (d *dummyOutput) Off() error {
	d.s = false
	return nil
}

func (*dummyInput) Close() error                     { return nil }
func (*dummyInput) AddCallback(DigitalInputCallback) {}

func NewNOOPDigitalInput() DigitalInput {
	return &dummyInput{}
}

func NewNOOPDigitalOutput() DigitalOutput {
	return &dummyOutput{}
}
