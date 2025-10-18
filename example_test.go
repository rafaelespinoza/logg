package logg_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/rafaelespinoza/logg"
)

// Setup the application logger with some preset fields and a logging destination.
// Then log.
func Example() {
	logg.Setup(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}),
		slog.String("branch_name", "main"),
		slog.String("build_time", "20060102T150415"),
		slog.String("commit_hash", "deadbeef"),
	)

	//
	// Log at the INFO level.
	//
	logg.Info("hello world")

	// Want message string interpolation? Do this beforehand.
	variable := "there"
	logg.Info(fmt.Sprintf("hello %s", variable))

	// Log message with attributes
	logg.Info("hello with attributes", slog.Bool("simple", true), slog.Int("rating", 42))

	//
	// Log at the ERROR level.
	//
	err := errors.New("example")
	logg.Error(err, "uh-oh")

	// Want message string interpolation? Do this beforehand.
	variable = "spaghetti-ohs"
	logg.Error(err, fmt.Sprintf("uh-oh, %s", variable))

	// Log message with attributes
	logg.Error(err, "uh-oh with attributes", slog.Bool("simple", true), slog.Int("rating", 42))
}

func ExampleSetup_textFormat() {
	logg.Setup(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}),
		slog.String("branch_name", "main"),
		slog.String("build_time", "20060102T150415"),
		slog.String("commit_hash", "deadbeef"),
	)
}

// Write to standard error as well as some file.
func ExampleSetup_multipleSinks() {
	var file io.Writer
	sink := io.MultiWriter(os.Stderr, file)

	logg.Setup(
		slog.NewJSONHandler(sink, &slog.HandlerOptions{Level: slog.LevelInfo}),
		slog.String("branch_name", "main"),
		slog.String("build_time", "20060102T150415"),
		slog.String("commit_hash", "deadbeef"),
	)
}

// Set an ID value and retrieve it. The tracing ID value is what you tell it.
// Derive a new context based on the previous context and set another ID.
func ExampleSetID() {
	const tracingID = "user-generated-random-value"
	ctxA := logg.SetID(context.Background(), tracingID)
	gotID, found := logg.GetID(ctxA)
	fmt.Printf("%q, %t\n", gotID, found)

	// Create a new context derived from the first context.
	const differentTracingID = "different-" + tracingID
	ctxB := logg.SetID(ctxA, differentTracingID)
	gotID, found = logg.GetID(ctxB)
	fmt.Printf("%q, %t\n", gotID, found)
	// Output:
	// "user-generated-random-value", true
	// "different-user-generated-random-value", true
}

// Set an ID value and then retrieve it. Retrieving from contexts without a set
// ID outputs zero values.
func ExampleGetID() {
	const tracingID = "user-generated-random-value"
	ctx := logg.SetID(context.Background(), tracingID)

	gotID, found := logg.GetID(ctx)
	fmt.Printf("%q, %t\n", gotID, found)

	// Demonstrate zero value.
	gotID, found = logg.GetID(context.Background())
	fmt.Printf("%q, %t\n", gotID, found)
	// Output:
	// "user-generated-random-value", true
	// "", false
}

// Initialize the logger with or without data fields.
func ExampleNew() {
	// An empty input handler makes a logger that writes to the same handler
	// provisioned in the Setup function. This logger has its own data fields.
	logger := logg.New(nil, slog.Bool("bravo", true), slog.Float64("foxtrot", 1.23), slog.Int("india", 10))
	logger.Info("hello, world")

	// This logger doesn't have any data fields.
	loggerNoFields := logg.New(nil)
	loggerNoFields.Info("no data fields here")

	// This logger writes to another Handler altogether, which might be useful
	// for testing.
	var testBuffer io.Writer
	handler := slog.NewJSONHandler(testBuffer, &slog.HandlerOptions{Level: slog.LevelInfo})
	loggerOwnHandler := logg.New(handler)
	loggerOwnHandler.Info("test")
}

func ExampleEmitter() {
	// Create a logger that writes to the package-level Handler, which is
	// configured via the Setup function.
	alfa := logg.New(nil)

	//
	// Log at the INFO level.
	//
	// Outputs an info log.
	alfa.Info("test")

	// Create a new Emitter with some attributes, then output an info log.
	bravo := alfa.WithData([]slog.Attr{slog.String("foo", "bar")})
	bravo.Info("test")

	// Outputs an info log with more data attributes.
	bravo.Info("test", slog.Bool("test", true))

	//
	// Log at the ERROR level.
	//
	err := errors.New("example")
	// Outputs an error log.
	alfa.Error(err, "test")

	// Outputs an error log with more data attributes.
	bravo.Error(err, "test", slog.Bool("test", true))
}

// Demonstrate tracing ID behavior. The tracing ID can be any string you want.
func ExampleEmitter_WithID() {
	// Create a new context to ensure the same tracing ID on each subsequent
	// tracing event.
	ctxA := logg.SetID(context.Background(), "AAA")
	alfa := logg.New(nil, slog.String("a", "A")).WithID(ctxA)
	// These events have a tracing ID AAA.
	alfa.Info("altoona")
	alfa.Info("alice in wonderland")

	// Use the same context in another Emitter.
	bravo := logg.New(nil, slog.String("b", "B")).WithID(ctxA)
	// These events have a tracing ID AAA.
	bravo.Info("brazil")
	bravo.Info("bilbo baggins")

	// Deriving a new context with its own tracing ID and passing it to another
	// Emitter yields the tracing ID from the derived context.
	anotherCtx := logg.SetID(ctxA, "CCC")
	charlie := logg.New(nil, slog.String("c", "C")).WithID(anotherCtx)
	// These events have a tracing ID CCC.
	charlie.Info("chicago")
	charlie.Info("chewbacca")
}
