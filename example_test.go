package logg_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/rafaelespinoza/logg"
)

// Configure the application logger with some preset fields and a logging destination.
func ExampleConfigure() {
	logg.Configure(
		os.Stderr,
		[]slog.Attr{
			slog.String("branch_name", "main"),
			slog.String("build_time", "20060102T150415"),
			slog.String("commit_hash", "deadbeef"),
		},
	)
}

// Write to standard error as well as some file.
func ExampleConfigure_multipleSinks() {
	var file io.Writer

	logg.Configure(
		os.Stderr,
		[]slog.Attr{
			slog.String("branch_name", "main"),
			slog.String("build_time", "20060102T150415"),
			slog.String("commit_hash", "deadbeef"),
		},
		file,
	)
}

// This example shows how the first usage of logg, that is not Configure,
// may unintentionally set up your logger.
func ExampleConfigure_possiblyUnintendedConfiguration() {
	logg.Info("these writes go")
	logg.Info("to standard error")
	logg.Info("by default")

	// Then your code attempts to configure the logger to write to someSocket.
	var someSocket io.Writer
	logg.Configure(
		someSocket,
		[]slog.Attr{
			slog.String("branch_name", "main"),
			slog.String("build_time", "20060102T150415"),
			slog.String("commit_hash", "deadbeef"),
		},
	)

	// Whoops, these logging event will continue to go to standard error. This
	// may not be what you want. The solution would be to call Configure before
	// emitting any kind of event.
	logg.Info("hello, is there")
	logg.Info("anybody out there?")
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

func ExampleError() {
	err := errors.New("example")
	logg.Error(err, "uh-oh")

	// Want message string interpolation? Do this beforehand.
	variable := "spaghetti-ohs"
	logg.Error(err, fmt.Sprintf("uh-oh, %s", variable))

	// Log message with attributes
	logg.Error(err, "uh-oh with attributes", slog.Bool("simple", true), slog.Int("rating", 42))
}

func ExampleInfo() {
	logg.Info("hello world")

	// Want message string interpolation? Do this beforehand.
	variable := "there"
	logg.Info(fmt.Sprintf("hello %s", variable))

	// Log message with attributes
	logg.Info("hello with attributes", slog.Bool("simple", true), slog.Int("rating", 42))
}

// Create a logger without any data fields.
func ExampleNew_noFields() {
	logger := logg.New(nil)

	// do stuff ...

	logger.Info("no data fields here")
}

// This logger will emit events to multiple destinations.
func ExampleNew_multipleSinks() {
	var file, socket io.Writer

	dataFields := []slog.Attr{
		slog.Bool("bravo", true),
		slog.Duration("delta", 234*time.Millisecond),
		slog.Float64("foxtrot", 1.23),
		slog.Int("india", 10),
	}
	logger := logg.New(dataFields, file, socket)

	// do stuff ...

	logger.Info("hello")
	logger.Info("world")
}

// Initialize the logger with data fields.
func ExampleNew_fields() {
	logger := logg.New([]slog.Attr{
		slog.Bool("bravo", true),
		slog.Duration("delta", 234*time.Millisecond),
		slog.Float64("foxtrot", 1.23),
		slog.Int("india", 10),
	})

	// do stuff ...

	logger.Info("hello")
	logger.Info("world")
}

// Set up a logger a tracing ID. The tracing ID can be any string you want.
func ExampleNew_withID() {
	ctx := logg.SetID(context.Background(), "logger_id")
	logger := logg.New([]slog.Attr{}).WithID(ctx)

	// do stuff ...

	logger.Info("hello")
	logger.Info("world")
}

// Demonstrate tracing ID behavior.
func ExampleEmitter_withID() {
	// Create a new context to ensure the same tracing ID on each subsequent
	// tracing event.
	ctxA := logg.SetID(context.Background(), "AAA")
	alfa := logg.New([]slog.Attr{slog.String("a", "A")}).WithID(ctxA)
	// These events have a tracing ID AAA.
	alfa.Info("altoona")
	alfa.Info("alice in wonderland")

	// Use the same context in another Emitter.
	bravo := logg.New([]slog.Attr{slog.String("b", "B")}).WithID(ctxA)
	// These events have a tracing ID AAA.
	bravo.Info("brazil")
	bravo.Info("bilbo baggins")

	// Deriving a new context with its own tracing ID and passing it to another
	// Emitter yields the tracing ID from the derived context.
	anotherCtx := logg.SetID(ctxA, "CCC")
	charlie := logg.New([]slog.Attr{slog.String("c", "C")}).WithID(anotherCtx)
	// These events have a tracing ID CCC.
	charlie.Info("chicago")
	charlie.Info("chewbacca")
}

func ExampleEmitter_error() {
	err := errors.New("example")

	alfa := logg.New([]slog.Attr{})

	// Outputs an error log.
	alfa.Error(err, "test")

	// Create a new Emitter with some attributes, then output an error log.
	bravo := alfa.WithData([]slog.Attr{slog.String("foo", "bar")})
	bravo.Error(err, "test")

	// Outputs an error log with more data attributes.
	bravo.Error(err, "test", slog.Bool("test", true))
}

func ExampleEmitter_info() {
	alfa := logg.New([]slog.Attr{})

	// Outputs an info log.
	alfa.Info("test")

	// Create a new Emitter with some attributes, then output an info log.
	bravo := alfa.WithData([]slog.Attr{slog.String("foo", "bar")})
	bravo.Info("test")

	// Outputs an info log with more data attributes.
	bravo.Info("test", slog.Bool("test", true))
}
