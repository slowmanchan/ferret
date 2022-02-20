package expressions

import (
	"context"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/events"
	"github.com/MontFerret/ferret/pkg/runtime/expressions/literals"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/MontFerret/ferret/pkg/runtime/values/types"
	"time"
)

type eventExpression struct {
	src            core.SourceMap
	eventName      core.Expression
	eventSource    core.Expression
	options        core.Expression
	timeout        core.Expression
	filterSrc      core.SourceMap
	filter         core.Expression
	filterVariable string
}

func newEventExpression(
	src core.SourceMap,
	eventName core.Expression,
	eventSource core.Expression,
) (*eventExpression, error) {
	if eventName == nil {
		return nil, core.Error(core.ErrInvalidArgument, "event name")
	}

	if eventSource == nil {
		return nil, core.Error(core.ErrMissedArgument, "event source")
	}

	return &eventExpression{
		src:         src,
		eventName:   eventName,
		eventSource: eventSource,
		timeout:     literals.NewIntLiteral(events.DefaultTimeout),
	}, nil
}

func (e *eventExpression) SetOptions(options core.Expression) error {
	if options == nil {
		return core.ErrInvalidArgument
	}

	e.options = options

	return nil
}

func (e *eventExpression) SetTimeout(timeout core.Expression) error {
	if timeout == nil {
		return core.ErrInvalidArgument
	}

	e.timeout = timeout

	return nil
}

func (e *eventExpression) SetFilter(src core.SourceMap, variable string, exp core.Expression) error {
	if variable == "" {
		return core.ErrInvalidArgument
	}

	if exp == nil {
		return core.ErrInvalidArgument
	}

	e.filterSrc = src
	e.filterVariable = variable
	e.filter = exp

	return nil
}

func (e *eventExpression) getEventName(ctx context.Context, scope *core.Scope) (values.String, error) {
	eventName, err := e.eventName.Exec(ctx, scope)

	if err != nil {
		return "", err
	}

	return values.ToString(eventName), nil
}

func (e *eventExpression) getOptions(ctx context.Context, scope *core.Scope) (*values.Object, error) {
	if e.options == nil {
		return nil, nil
	}

	options, err := e.options.Exec(ctx, scope)

	if err != nil {
		return nil, err
	}

	if err := core.ValidateType(options, types.Object); err != nil {
		return nil, err
	}

	return options.(*values.Object), nil
}

func (e *eventExpression) getTimeout(ctx context.Context, scope *core.Scope) (time.Duration, error) {
	timeoutValue, err := e.timeout.Exec(ctx, scope)

	if err != nil {
		return 0, err
	}

	timeoutInt := values.ToIntDefault(timeoutValue, events.DefaultTimeout)

	return time.Duration(timeoutInt) * time.Millisecond, nil
}
