package expressions

import (
	"context"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/events"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/pkg/errors"
)

type CreateEventExpression struct {
	*eventExpression
	eventArgs core.Expression
}

func NewCreateEventExpression(
	src core.SourceMap,
	eventName core.Expression,
	eventSource core.Expression,
) (*CreateEventExpression, error) {
	base, err := newEventExpression(src, eventName, eventSource)

	if err != nil {
		return nil, err
	}

	return &CreateEventExpression{
		eventExpression: base,
	}, nil
}

func (e *CreateEventExpression) SetArguments(args core.Expression) error {
	if args == nil {
		return core.ErrInvalidArgument
	}

	e.eventArgs = args

	return nil
}

func (e *CreateEventExpression) Exec(ctx context.Context, scope *core.Scope) (core.Value, error) {
	eventSource, err := e.eventSource.Exec(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, err)
	}

	dispatcher, ok := eventSource.(events.Dispatcher)

	if !ok {
		return values.None, core.TypeError(eventSource.Type(), core.NewType("Dispatcher"))
	}

	eventName, err := e.getEventName(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, errors.Wrap(err, "unable to calculate event name"))
	}

	eventArgs, err := e.getArguments(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, errors.Wrap(err, "unable to get event arguments"))
	}

	opts, err := e.getOptions(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, err)
	}

	timeout, err := e.getTimeout(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, errors.Wrap(err, "failed to calculate timeout value"))
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return dispatcher.Dispatch(ctx, events.Action{
		Name:    eventName,
		Args:    eventArgs,
		Options: opts,
	})
}

func (e *CreateEventExpression) getArguments(ctx context.Context, scope *core.Scope) (core.Value, error) {
	if e.eventArgs == nil {
		return values.None, nil
	}

	return e.eventArgs.Exec(ctx, scope)
}
