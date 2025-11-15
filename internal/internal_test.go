package internal_test

import (
	"log/slog"
	"testing"

	"github.com/rafaelespinoza/logg/internal"
	st "github.com/rafaelespinoza/logg/slogtesting"
)

func TestSlogGroupAttrs(t *testing.T) {
	got := internal.SlogGroupAttrs("k", slog.String("foo", "bar"), slog.Int("i", 1))

	attrs := []slog.Attr{got}
	checkGroupAttrs := st.InGroup("k",
		st.HasAttr(slog.String("foo", "bar")),
		st.HasAttr(slog.Int("i", 1)),
	)
	if err := checkGroupAttrs(attrs); err != nil {
		t.Fatal(err)
	}
}
