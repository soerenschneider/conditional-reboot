package preconditions

import (
	"errors"
	"time"
)

const WindowedPreconditionName = "time_window"

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (r *realClock) Now() time.Time {
	return time.Now()
}

type WindowedPrecondition struct {
	FromHour int
	ToHour   int
	clock    Clock
}

func NewWindowPrecondition(fromHour, toHour int) (*WindowedPrecondition, error) {
	if fromHour < 0 || fromHour > 23 {
		return nil, errors.New("fromHour must not be in the interval [0, 23]")
	}

	if toHour < 0 || toHour > 23 {
		return nil, errors.New("toHour must not be in the interval [0, 23]")
	}

	if fromHour == toHour {
		return nil, errors.New("fromHour and toHour must not be identical")
	}

	return &WindowedPrecondition{
		FromHour: fromHour,
		ToHour:   toHour,
		clock:    &realClock{},
	}, nil
}

func WindowPreconditionFromMap(args map[string]any) (*WindowedPrecondition, error) {
	if args == nil {
		return nil, errors.New("empty args provided")
	}

	intFrom, ok := args["from"].(float64)
	if !ok {
		return nil, errors.New("can not cast 'from")
	}

	intTo, ok := args["to"].(float64)
	if !ok {
		return nil, errors.New("can not cast 'to")
	}

	return NewWindowPrecondition(int(intFrom), int(intTo))
}

func (c *WindowedPrecondition) PerformCheck() bool {
	now := c.clock.Now()
	currentHour := now.Hour()

	// handle overnight ranges
	if c.FromHour > c.ToHour {
		c.ToHour += 24
		currentHour += 24
	}

	return currentHour >= c.FromHour && currentHour < c.ToHour
}
