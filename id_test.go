package logg_test

import (
	"context"
	"testing"

	"github.com/rafaelespinoza/logg"
)

func TestID(t *testing.T) {
	const testID = "alfa"
	ctxA := logg.SetID(context.Background(), testID)

	gotID, found := logg.GetID(ctxA)
	if !found {
		t.Fatal("expected for context to have a tracing ID")
	} else if gotID != testID {
		t.Fatalf("wrong value for tracing ID; got %q, expected %q", gotID, testID)
	}

	// Make a new context with its own ID.
	const differentTestID = "bravo"
	ctxB := logg.SetID(ctxA, differentTestID)

	gotID, found = logg.GetID(ctxB)
	if !found {
		t.Fatal("expected for context to have a tracing ID")
	} else if gotID != differentTestID {
		t.Fatalf("wrong id, got %q, expected %q", gotID, differentTestID)
	}

	// Deriving a second context does not change the ID on the first context.
	if gotID, found = logg.GetID(ctxA); !found {
		t.Fatal("expected for context to have a tracing ID")
	} else if gotID != testID {
		t.Fatalf("wrong value for tracing ID; got %q, expected %q", gotID, testID)
	}
}
