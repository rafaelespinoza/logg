package internal_test

import (
	"log/slog"
	"testing"

	"github.com/rafaelespinoza/logg/internal"
)

func TestSlogGroupAttrs(t *testing.T) {
	got := internal.SlogGroupAttrs("k", slog.String("foo", "bar"), slog.Int("i", 1))

	if k := got.Value.Kind(); k != slog.KindGroup {
		t.Fatalf("wrong kind; got %q, expected %q", k.String(), slog.KindGroup.String())
	}
}
