package logg_test

import (
	"errors"
	"testing"

	"github.com/rafaelespinoza/logg"
)

func TestLogger(t *testing.T) {
	// These tests mostly check that constructing a Logger with various input
	// combinations can work without panicking on invalid memory address refs.

	t.Run("no sinks", func(t *testing.T) {
		logg.New(nil).Infof(t.Name())
		logg.New(map[string]interface{}{"a": "b"}).Infof(t.Name())
	})

	t.Run("empty sinks", func(t *testing.T) {
		logg.New(nil, nil).Infof(t.Name())
		logg.New(map[string]interface{}{"a": "b"}, nil).Infof(t.Name())
	})

	t.Run("one sink", func(t *testing.T) {
		alfa := newDataSink()
		logg.New(nil, alfa).Infof(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		bravo := newDataSink()
		logg.New(map[string]interface{}{"a": "b"}, bravo).Infof(t.Name())
		if len(bravo.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("more sinks", func(t *testing.T) {
		alfa, bravo := newDataSink(), newDataSink()
		logg.New(nil, alfa, bravo).Infof(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
		if len(bravo.Raw()) < 1 {
			t.Error("did not write data")
		}

		charlie, delta := newDataSink(), newDataSink()
		logg.New(map[string]interface{}{"a": "b"}, charlie, delta).Infof(t.Name())
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
		logg.New(nil, alfa).Infof(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		logg.New(nil, alfa).Errorf(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to WARN", func(t *testing.T) {
		t.Setenv("LOGG_LEVEL", "WARN")

		alfa := newDataSink()
		logg.New(nil, alfa).Infof(t.Name())
		if len(alfa.Raw()) > 0 {
			t.Error("unexpected data written for current logging level")
		}

		logg.New(nil, alfa).Errorf(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to ERROR", func(t *testing.T) {
		t.Setenv("LOGG_LEVEL", "ERROR")

		alfa := newDataSink()
		logg.New(nil, alfa).Infof(t.Name())
		if len(alfa.Raw()) > 0 {
			t.Error("unexpected data written for current logging level")
		}

		logg.New(nil, alfa).Errorf(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("log level set to unknown logging level", func(t *testing.T) {
		// Check that the library works despite unknown setting.
		t.Setenv("LOGG_LEVEL", "UNKNOWN")

		alfa := newDataSink()
		logg.New(nil, alfa).Infof(t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}

		logg.New(nil, alfa).Errorf(errors.New("test"), t.Name())
		if len(alfa.Raw()) < 1 {
			t.Error("did not write data")
		}
	})
}
