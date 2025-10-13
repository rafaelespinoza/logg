package logg_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/rafaelespinoza/logg"
)

func TestLogger(t *testing.T) {
	// These tests mostly check that constructing a Logger with various input
	// combinations can work without panicking on invalid memory address refs.

	t.Run("no sinks", func(t *testing.T) {
		logg.New(nil).Info(t.Name())
		logg.New([]slog.Attr{slog.String("a", "b")}).Info(t.Name())
	})

	t.Run("empty sinks", func(t *testing.T) {
		logg.New(nil, nil).Info(t.Name())
		logg.New([]slog.Attr{slog.String("a", "b")}, nil).Info(t.Name())
	})

	t.Run("one sink", func(t *testing.T) {
		alfa := newDataSink()
		logg.New(nil, alfa).Info(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		bravo := newDataSink()
		logg.New([]slog.Attr{slog.String("a", "b")}, bravo).Info(t.Name())
		if len(bravo.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("more sinks", func(t *testing.T) {
		alfa, bravo := newDataSink(), newDataSink()
		logg.New(nil, alfa, bravo).Info(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
		if len(bravo.Raw()) < 1 {
			t.Error("did not write data")
		}

		charlie, delta := newDataSink(), newDataSink()
		logg.New([]slog.Attr{slog.String("a", "b")}, charlie, delta).Info(t.Name())
		if len(charlie.Raw()) < 1 {
			t.Error("did not write data")
		}
		if len(delta.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to INFO", func(t *testing.T) {
		t.Setenv("LOGG_LEVEL", "INFO")

		alfa := newDataSink()
		logg.New(nil, alfa).Info(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		logg.New(nil, alfa).Error(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to WARN", func(t *testing.T) {
		t.Setenv("LOGG_LEVEL", "WARN")

		alfa := newDataSink()
		logg.New(nil, alfa).Info(t.Name())
		if len(alfa.Raw()) > 0 {
			t.Error("unexpected data written for current logging level")
		}

		logg.New(nil, alfa).Error(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to ERROR", func(t *testing.T) {
		t.Setenv("LOGG_LEVEL", "ERROR")

		alfa := newDataSink()
		logg.New(nil, alfa).Info(t.Name())
		if len(alfa.Raw()) > 0 {
			t.Error("unexpected data written for current logging level")
		}

		logg.New(nil, alfa).Error(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to unknown logging level", func(t *testing.T) {
		// Check that the library works despite unknown setting.
		t.Setenv("LOGG_LEVEL", "UNKNOWN")

		alfa := newDataSink()
		logg.New(nil, alfa).Info(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		logg.New(nil, alfa).Error(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})
}
