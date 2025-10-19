package logg_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/rafaelespinoza/logg"
)

func TestLogger(t *testing.T) {
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
