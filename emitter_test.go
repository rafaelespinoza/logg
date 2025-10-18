package logg_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/rafaelespinoza/logg"
)

func TestEmitter(t *testing.T) {
	// These tests mostly check that constructing a Logger with various input
	// combinations can work without panicking on invalid memory address refs.

	t.Run("empty sinks", func(t *testing.T) {
		logg.New(nil).Info(t.Name())
		logg.New(nil, slog.String("a", "b")).Info(t.Name())
	})

	t.Run("one sink", func(t *testing.T) {
		alfa, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.New(handler).Info(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		bravo, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.New(handler, slog.String("a", "b")).Info(t.Name())
		if len(bravo.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("more sinks", func(t *testing.T) {
		alfa, bravo := newDataSink(), newDataSink()
		handlerAB := slog.NewJSONHandler(io.MultiWriter(alfa, bravo), &slog.HandlerOptions{Level: slog.LevelInfo})
		logg.New(handlerAB).Info(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
		if len(bravo.Raw()) < 1 {
			t.Error("did not write data")
		}

		charlie, delta := newDataSink(), newDataSink()
		handlerCD := slog.NewJSONHandler(io.MultiWriter(charlie, delta), &slog.HandlerOptions{Level: slog.LevelInfo})
		logg.New(handlerCD, slog.String("a", "b")).Info(t.Name())
		if len(charlie.Raw()) < 1 {
			t.Error("did not write data")
		}
		if len(delta.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("handler level", func(t *testing.T) {
		tests := []struct {
			inputLevel     slog.Level
			expDebugOutput bool
			expInfoOutput  bool
			expWarnOutput  bool
			expErrorOutput bool
		}{
			{
				inputLevel:     slog.LevelDebug,
				expDebugOutput: true, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelInfo,
				expDebugOutput: false, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelWarn,
				expDebugOutput: false, expInfoOutput: false, expWarnOutput: true, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelError,
				expDebugOutput: false, expInfoOutput: false, expWarnOutput: false, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelError + 1,
				expDebugOutput: false, expInfoOutput: false, expWarnOutput: false, expErrorOutput: false,
			},
		}

		for _, test := range tests {
			t.Run(test.inputLevel.String(), func(t *testing.T) {
				sink, handler := newDataSinkAndJSONHandler(test.inputLevel)

				logg.New(handler).Debug(t.Name())
				gotDebug := sink.Raw()
				if test.expDebugOutput && len(gotDebug) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelDebug.String())
				} else if !test.expInfoOutput && len(gotDebug) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelDebug.String())
				}

				logg.New(handler).Info(t.Name())
				gotInfo := sink.Raw()
				if test.expInfoOutput && len(gotInfo) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelInfo.String())
				} else if !test.expInfoOutput && len(gotInfo) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelInfo.String())
				}

				logg.New(handler).Warn(t.Name())
				gotWarn := sink.Raw()
				if test.expWarnOutput && len(gotWarn) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelWarn.String())
				} else if !test.expWarnOutput && len(gotWarn) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelWarn.String())
				}

				logg.New(handler).Error(errors.New("test"), t.Name())
				gotError := sink.Raw()
				if test.expErrorOutput && len(gotError) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelError.String())
				} else if !test.expErrorOutput && len(gotError) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelError.String())
				}
			})
		}
	})
}

func TestEmitterDebug(t *testing.T) {
	setupPackageVars(pkgSink)

	t.Run("from logger", func(t *testing.T) {
		// setup
		sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		// action
		logger.Debug("hello")

		// test
		testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("from logger.WithData", func(t *testing.T) {
		// setup
		sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		// action
		logger.WithData([]slog.Attr{
			slog.Bool("bravo", true),
			slog.String("delta", (234 * time.Millisecond).String()),
			slog.Float64("foxtrot", 1.23),
			slog.Int("india", 10),
		}).Debug("goodbye")

		// test
		testLogg(t, sink.Raw(), nil, "goodbye", false, "", []slog.Attr{
			slog.Bool("bravo", true),
			slog.String("delta", (234 * time.Millisecond).String()),
			slog.Float64("foxtrot", 1.23),
			slog.Int("india", 10),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("another event from same logger can have its own fields", func(t *testing.T) {
		// setup
		sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
		logger := logg.New(handler, slog.String("sierra", "nevada"))
		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Debug("goodbye")

		// action
		logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Debug("goodbye again")

		// test
		testLogg(t, sink.Raw(), nil, "goodbye again", false, "", []slog.Attr{
			slog.Bool("zulu", true),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		// setup
		sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		// action
		logger.Debug("hello", slog.Int("i", 1))

		// test
		testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{
			slog.String("sierra", "nevada"),
			slog.Int("i", 1),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// Calling a log method with attrs does not affect logger's fields
		logger.Debug("hello again")

		testLogg(t, sink.Raw(), nil, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("not enabled", func(t *testing.T) {
		// setup
		sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
		logger := logg.New(handler)

		// action
		logger.Debug("hello")

		// test
		if len(sink.Raw()) > 0 {
			t.Errorf("unexpected data written at level %s", slog.LevelDebug.String())
		}
	})
}

func TestEmitterInfo(t *testing.T) {
	setupPackageVars(pkgSink)

	t.Run("from logger", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.Info("hello")

		testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("from logger.WithData", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.WithData([]slog.Attr{
			slog.Bool("bravo", true),
			slog.String("delta", (234 * time.Millisecond).String()),
			slog.Float64("foxtrot", 1.23),
			slog.Int("india", 10),
		}).Info("goodbye")

		testLogg(t, sink.Raw(), nil, "goodbye", false, "", []slog.Attr{
			slog.Bool("bravo", true),
			slog.String("delta", (234 * time.Millisecond).String()),
			slog.Float64("foxtrot", 1.23),
			slog.Int("india", 10),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("another event from same logger can have its own fields", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))
		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Info("goodbye")

		logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Info("goodbye again")

		testLogg(t, sink.Raw(), nil, "goodbye again", false, "", []slog.Attr{
			slog.Bool("zulu", true),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.Info("hello", slog.Int("i", 1))

		testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{
			slog.String("sierra", "nevada"),
			slog.Int("i", 1),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// Calling a log method with attrs does not affect logger's fields
		logger.Info("hello again")

		testLogg(t, sink.Raw(), nil, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("not enabled", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
		logger := logg.New(handler)

		logger.Info("hello")

		if len(sink.Raw()) > 0 {
			t.Errorf("unexpected data written at level %s", slog.LevelInfo.String())
		}
	})
}

func TestEmitterWarn(t *testing.T) {
	setupPackageVars(pkgSink)

	t.Run("from logger", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.Warn("hello")

		testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("from logger.WithData", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.WithData([]slog.Attr{
			slog.Bool("bravo", true),
			slog.String("delta", (234 * time.Millisecond).String()),
			slog.Float64("foxtrot", 1.23),
			slog.Int("india", 10),
		}).Warn("goodbye")

		testLogg(t, sink.Raw(), nil, "goodbye", false, "", []slog.Attr{
			slog.Bool("bravo", true),
			slog.String("delta", (234 * time.Millisecond).String()),
			slog.Float64("foxtrot", 1.23),
			slog.Int("india", 10),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("another event from same logger can have its own fields", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))
		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Warn("goodbye")

		logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Warn("goodbye again")

		testLogg(t, sink.Raw(), nil, "goodbye again", false, "", []slog.Attr{
			slog.Bool("zulu", true),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.Warn("hello", slog.Int("i", 1))

		testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{
			slog.String("sierra", "nevada"),
			slog.Int("i", 1),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// Calling a log method with attrs does not affect logger's fields
		logger.Warn("hello again")

		testLogg(t, sink.Raw(), nil, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("not enabled", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
		logger := logg.New(handler)

		logger.Warn("hello")

		if len(sink.Raw()) > 0 {
			t.Errorf("unexpected data written at level %s", slog.LevelWarn.String())
		}
	})
}

func TestEmitterError(t *testing.T) {
	setupPackageVars(pkgSink)

	t.Run("from logger", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.Error(errors.New("hello"), "logger error")

		testLogg(t, sink.Raw(), errors.New("hello"), "logger error", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("from logger.WithData", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.WithData([]slog.Attr{
			slog.Bool("bravo", false),
			slog.String("delta", (432 * time.Millisecond).String()),
			slog.Float64("foxtrot", 3.21),
			slog.Int("india", 100),
		}).Error(errors.New("goodbye"), "event error")

		testLogg(t, sink.Raw(), errors.New("goodbye"), "event error", false, "", []slog.Attr{
			slog.Bool("bravo", false),
			slog.String("delta", (432 * time.Millisecond).String()),
			slog.Float64("foxtrot", 3.21),
			slog.Int("india", 100),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("another event from same logger can have its own fields", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))
		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Info("goodbye")

		logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Error(errors.New("bye"), "goodbye again")

		testLogg(t, sink.Raw(), errors.New("bye"), "goodbye again", false, "", []slog.Attr{
			slog.Bool("zulu", true),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		err := errors.New("test")
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("sierra", "nevada"))

		logger.Error(err, "hello", slog.Int("i", 1))

		testLogg(t, sink.Raw(), err, "hello", false, "", []slog.Attr{
			slog.String("sierra", "nevada"),
			slog.Int("i", 1),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// Calling a log method with attrs does not affect logger's fields
		logger.Error(err, "hello again")

		testLogg(t, sink.Raw(), err, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("not enabled", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
		logger := logg.New(handler)

		logger.Error(errors.New("test"), "hello")

		if len(sink.Raw()) > 0 {
			t.Errorf("unexpected data written at level %s", slog.LevelError.String())
		}
	})
}

func TestEmitterWithID(t *testing.T) {
	// In these tests, a logger is referred to as an *Emitter value constructed
	// via New. An event is referred to as an inline chain of method calls
	// ending in a logging output method, such as Info.

	t.Run("logger ID is present in event", func(t *testing.T) {
		// logger calls WithID but event does not
		ctx := logg.SetID(context.Background(), "logger_id")
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler).WithID(ctx)
		logger.Info("logger with id")
		testLogg(t, sink.Raw(), nil, "logger with id", true, "logger_id", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		logger.Info("event with id")
		testLogg(t, sink.Raw(), nil, "event with id", true, "logger_id", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("event can set its own ID", func(t *testing.T) {
		// The logger does not call WithID, but the event does.
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler)
		logger.Info("logger without id")
		testLogg(t, sink.Raw(), nil, "logger without id", false, "", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		ctx := logg.SetID(context.Background(), "event_id")
		logger.WithID(ctx).Info("event with own id")
		testLogg(t, sink.Raw(), nil, "event with own id", true, "event_id", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("event can set its own ID despite logger setting an ID", func(t *testing.T) {
		// The logger calls WithID and so does the event.
		ctxA := logg.SetID(context.Background(), "logger_id")
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler).WithID(ctxA)
		logger.Info("logger with id")
		testLogg(t, sink.Raw(), nil, "logger with id", true, "logger_id", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// derive context from previous context, but set another ID.
		ctxB := logg.SetID(ctxA, "event_id")
		logger.WithID(ctxB).Info("event with own id")
		testLogg(t, sink.Raw(), nil, "event with own id", true, "event_id", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestEmitterWithData(t *testing.T) {
	t.Run("when initial state of fields is nil", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler).WithData([]slog.Attr{slog.String("foo", "alfa")})
		logger.Info("a")
		testLogg(t, sink.Raw(), nil, "a", false, "", []slog.Attr{slog.String("foo", "alfa")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("when passed nil", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler, slog.String("foo", "alfa")).WithData(nil)
		logger.Info("a")
		testLogg(t, sink.Raw(), nil, "a", false, "", []slog.Attr{slog.String("foo", "alfa")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("allows input fields to replace existing fields", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		loggerA := logg.New(handler, slog.String("foo", "alfa"), slog.Bool("bar", true))
		loggerA.Info("a")
		testLogg(t, sink.Raw(), nil, "a", false, "", []slog.Attr{slog.String("foo", "alfa"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// create a logger based on another logger.
		loggerB := loggerA.WithData([]slog.Attr{slog.String("foo", "bravo")})
		loggerB.Info("b")
		testLogg(t, sink.Raw(), nil, "b", false, "", []slog.Attr{slog.String("foo", "bravo"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// create another logger based on another logger which is itself based on another logger.
		loggerC := loggerB.WithData([]slog.Attr{slog.Bool("bar", false)})
		loggerC.Info("c")
		testLogg(t, sink.Raw(), nil, "c", false, "", []slog.Attr{slog.String("foo", "bravo"), slog.Bool("bar", false)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// check that loggerA's original data hasn't changed unexpectedly.
		loggerA.Info("d")
		testLogg(t, sink.Raw(), nil, "d", false, "", []slog.Attr{slog.String("foo", "alfa"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// check that loggerB's original data hasn't changed unexpectedly.
		loggerB.Info("e")
		testLogg(t, sink.Raw(), nil, "e", false, "", []slog.Attr{slog.String("foo", "bravo"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("duplicate key in existing attributes", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.Bool("bar", true), slog.Bool("bar", false)).
			WithData([]slog.Attr{slog.String("foo", "alfa")})

		logger.Info("test")
		testLogg(t, sink.Raw(), nil, "test", false, "", []slog.Attr{slog.String("foo", "alfa"), slog.Bool("bar", false)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("duplicate key in next attributes", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logger := logg.New(handler, slog.String("foo", "alfa")).
			WithData([]slog.Attr{slog.Bool("bar", true), slog.Bool("bar", false)})

		logger.Info("test")
		testLogg(t, sink.Raw(), nil, "test", false, "", []slog.Attr{slog.String("foo", "alfa"), slog.Bool("bar", false)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}
