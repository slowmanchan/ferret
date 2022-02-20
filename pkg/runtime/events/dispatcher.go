package events

import (
	"context"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
)

type (
	Action struct {
		Name    values.String
		Args    core.Value
		Options *values.Object
	}

	Dispatcher interface {
		Dispatch(ctx context.Context, action Action) (core.Value, error)
	}
)
