package logg

import (
	"context"
	"testing"
)

func TestID(t *testing.T) {
	// Create a new ID.
	ctx, got := getSetID(context.Background())
	if got == "" {
		t.Fatal("id should be non-empty")
	}
	exp := got

	// Attempting to create another ID with same context outputs the same data.
	ctx, got = getSetID(ctx)
	if got != exp {
		t.Errorf("wrong id, got %q, expected %q", got, exp)
	}
}
