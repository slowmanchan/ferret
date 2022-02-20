package expressions

import (
	"context"
	"github.com/pkg/errors"

	"github.com/MontFerret/ferret/pkg/runtime/collections"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/events"
	"github.com/MontFerret/ferret/pkg/runtime/expressions/clauses"
	"github.com/MontFerret/ferret/pkg/runtime/values"
)

type WaitForEventExpression struct {
	*eventExpression
}

func NewWaitForEventExpression(
	src core.SourceMap,
	eventName core.Expression,
	eventSource core.Expression,
) (*WaitForEventExpression, error) {
	base, err := newEventExpression(src, eventName, eventSource)

	if err != nil {
		return nil, err
	}

	return &WaitForEventExpression{
		eventExpression: base,
	}, nil
}

func (e *WaitForEventExpression) Exec(ctx context.Context, scope *core.Scope) (core.Value, error) {
	eventName, err := e.getEventName(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, errors.Wrap(err, "unable to calculate event name"))
	}

	eventSource, err := e.eventSource.Exec(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, err)
	}

	observable, ok := eventSource.(events.Observable)

	if !ok {
		return values.None, core.TypeError(eventSource.Type(), core.NewType("Observable"))
	}

	opts, err := e.getOptions(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, err)
	}

	timeout, err := e.getTimeout(ctx, scope)

	if err != nil {
		return values.None, core.SourceError(e.src, errors.Wrap(err, "failed to calculate timeout value"))
	}

	subscription := events.Subscription{
		EventName: eventName,
		Options:   opts,
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var val core.Value

	if e.filter == nil {
		val, err = e.consumeFirst(ctx, observable, subscription)
	} else {
		val, err = e.consumeFiltered(ctx, scope, observable, subscription)
	}

	if err != nil {
		return nil, core.SourceError(e.src, err)
	}

	return val, nil
}

func (e *WaitForEventExpression) consumeFirst(ctx context.Context, observable events.Observable, subscription events.Subscription) (core.Value, error) {
	stream, err := observable.Subscribe(ctx, subscription)

	if err != nil {
		return values.None, err
	}

	defer stream.Close(ctx)

	select {
	case evt := <-stream.Read(ctx):
		if err := evt.Err(); err != nil {
			return values.None, err
		}

		return evt.Value(), nil
	case <-ctx.Done():
		return values.None, ctx.Err()
	}
}

func (e *WaitForEventExpression) consumeFiltered(ctx context.Context, scope *core.Scope, observable events.Observable, subscription events.Subscription) (core.Value, error) {
	stream, err := observable.Subscribe(ctx, subscription)

	if err != nil {
		return values.None, err
	}

	defer stream.Close(ctx)

	iterable, err := clauses.NewFilterClause(
		e.filterSrc,
		collections.AsIterable(func(c context.Context, scope *core.Scope) (collections.Iterator, error) {
			return collections.FromCoreIterator(e.filterVariable, "", events.NewIterator(stream.Read(c)))
		}),
		e.filter,
	)

	if err != nil {
		return values.None, err
	}

	iter, err := iterable.Iterate(ctx, scope)

	if err != nil {
		return values.None, err
	}

	out, err := iter.Next(ctx, scope)

	if err != nil {
		return values.None, err
	}

	return out.GetVariable(e.filterVariable)
}
