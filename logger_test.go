package logg_test

import (
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
}
