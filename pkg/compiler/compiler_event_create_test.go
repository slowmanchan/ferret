package compiler_test

import (
	"context"
	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/runtime"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/events"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type TestDispatcher struct {
	*values.Object

	actions []events.Action
}

func NewTestDispatcher() *TestDispatcher {
	return &TestDispatcher{
		Object:  values.NewObject(),
		actions: make([]events.Action, 0, 10),
	}
}

func (t *TestDispatcher) Assert(action events.Action, expected int) error {
	actual := 0

	for _, a := range t.actions {
		if a.Name == action.Name {
			if values.Compare(a.Args, action.Args) == 0 &&
				values.Compare(a.Options, action.Options) == 0 {
				actual++
			}
		}
	}

	if expected != actual {
		return errors.Errorf("%s expected to be dispatched %d, but actual count is %d", action.Name, expected, actual)
	}

	return nil
}

func (t *TestDispatcher) Dispatch(_ context.Context, action events.Action) (core.Value, error) {
	t.actions = append(t.actions, action)

	return values.NewInt(len(t.actions)), nil
}

func newCompilerWithDispatcher() (*compiler.Compiler, *TestDispatcher) {
	c := compiler.New()
	d := NewTestDispatcher()

	err := c.Namespace("X").
		RegisterFunctions(core.NewFunctionsFromMap(
			map[string]core.Function{
				"DISPATCHER": func(ctx context.Context, args ...core.Value) (core.Value, error) {
					return d, nil
				},
			},
		))

	So(err, ShouldBeNil)

	return c, d
}

func TestCreateEventExpression(t *testing.T) {
	Convey("CREATE EVENT parser", t, func() {
		Convey("Should parse", func() {
			c, _ := newCompilerWithDispatcher()

			_, err := c.Compile(`
LET obj = {}

CREATE EVENT "test" IN obj

RETURN NONE
`)

			So(err, ShouldBeNil)
		})

		Convey("Should parse 2", func() {
			c, _ := newCompilerWithDispatcher()

			_, err := c.Compile(`
LET obj = {}

CREATE EVENT "test" IN obj TIMEOUT 1000

RETURN NONE
`)

			So(err, ShouldBeNil)
		})

		Convey("Should parse 3", func() {
			c, _ := newCompilerWithDispatcher()

			_, err := c.Compile(`
LET obj = {}

LET tmt = 1000
CREATE EVENT "test" IN obj TIMEOUT tmt

RETURN NONE
`)

			So(err, ShouldBeNil)
		})

		Convey("Should parse 4", func() {
			c, _ := newCompilerWithDispatcher()

			_, err := c.Compile(`
LET obj = {}

LET tmt = 1000
CREATE EVENT "test" IN obj TIMEOUT tmt

RETURN NONE
`)

			So(err, ShouldBeNil)
		})
	})

	Convey("CREATE EVENT X IN Y runtime", t, func() {
		Convey("Should wait for a given event", func() {
			c, d := newCompilerWithDispatcher()

			prog := c.MustCompile(`
LET obj = X::DISPATCHER()

CREATE EVENT "test" IN obj

RETURN NONE
`)

			_, err := prog.Run(context.Background())

			So(err, ShouldBeNil)
			So(d.Assert(events.Action{Name: "test"}, 1), ShouldBeNil)
		})

		Convey("Should wait for a given event using variable", func() {
			c, d := newCompilerWithDispatcher()

			prog := c.MustCompile(`
LET eventName = "test"
LET obj = X::DISPATCHER()

CREATE EVENT eventName IN obj

RETURN NONE
`)

			_, err := prog.Run(context.Background())

			So(err, ShouldBeNil)

			So(d.Assert(events.Action{Name: "test"}, 1), ShouldBeNil)
		})

		Convey("Should wait for a given event using object", func() {
			c, d := newCompilerWithDispatcher()

			prog := c.MustCompile(`
LET evt = {
   name: "test"
}
LET obj = X::DISPATCHER()

CREATE EVENT evt.name IN obj
CREATE EVENT evt.name IN obj

RETURN NONE
`)

			_, err := prog.Run(context.Background())

			So(err, ShouldBeNil)
			So(d.Assert(events.Action{Name: "test"}, 2), ShouldBeNil)
		})

		Convey("Should wait for a given event using param", func() {
			c, d := newCompilerWithDispatcher()

			prog := c.MustCompile(`
LET obj = X::DISPATCHER()

CREATE EVENT @evt IN obj

RETURN NONE
`)

			_, err := prog.Run(context.Background(), runtime.WithParam("evt", "test"))

			So(err, ShouldBeNil)
			So(d.Assert(events.Action{Name: "test"}, 1), ShouldBeNil)
		})

		Convey("Should use arguments", func() {
			c, d := newCompilerWithDispatcher()

			prog := c.MustCompile(`
			LET obj = X::DISPATCHER()

			CREATE EVENT "test" IN obj WITH "foo"
			CREATE EVENT "test2" IN obj WITH ["foo", "bar"]
			
			RETURN NONE
			`)

			_, err := prog.Run(context.Background())

			So(err, ShouldBeNil)

			So(d.Assert(events.Action{Name: "test", Args: values.NewString("foo")}, 1), ShouldBeNil)
			So(d.Assert(events.Action{Name: "test2", Args: values.NewArrayWith(values.NewString("foo"), values.NewString("bar"))}, 1), ShouldBeNil)
		})

		Convey("Should use options", func() {
			c, d := newCompilerWithDispatcher()

			prog := c.MustCompile(`
					LET obj = X::DISPATCHER()
		
					CREATE EVENT "test" IN obj OPTIONS { value: "foo" }
		
					RETURN NONE
					`)

			_, err := prog.Run(context.Background())

			So(err, ShouldBeNil)

			So(d.Assert(events.Action{Name: "test", Options: values.NewObjectWith(values.NewObjectProperty("value", values.NewString("foo")))}, 1), ShouldBeNil)
		})
	})

}
