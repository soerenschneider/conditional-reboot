package checkers

import (
	"context"
)

type Checker interface {
	IsHealthy(ctx context.Context) (bool, error)
	Name() string
}
