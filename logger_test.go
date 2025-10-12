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

	t.Run("log level set to INFO", func(t *testing.T) {
		const level = slog.LevelInfo
		sink, handler := newDataSinkAndJSONHandler(level)

		logg.New(handler).Info(t.Name())
		if len(sink.Raw()) < 1 {
			t.Error("did not write data")
		}

		logg.New(handler).Error(errors.New("test"), t.Name())
		if len(sink.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to WARN", func(t *testing.T) {
		const level = slog.LevelWarn
		sink, handler := newDataSinkAndJSONHandler(level)

		logg.New(handler).Info(t.Name())
		if len(sink.Raw()) > 0 {
			t.Error("unexpected data written for current logging level")
		}

		logg.New(handler).Error(errors.New("test"), t.Name())
		if len(sink.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to ERROR", func(t *testing.T) {
		const level = slog.LevelError
		sink, handler := newDataSinkAndJSONHandler(level)

		logg.New(handler).Info(t.Name())
		if len(sink.Raw()) > 0 {
			t.Error("unexpected data written for current logging level")
		}

		logg.New(handler).Error(errors.New("test"), t.Name())
		if len(sink.Raw()) < 1 {
			t.Error("did not write data")
		}
	})
}
