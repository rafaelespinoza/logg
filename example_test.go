package logg_test

import (
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
	logg.SetDefaults(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}),
		&logg.Settings{
			ApplicationMetadata: []slog.Attr{
				slog.String("branch_name", "main"),
				slog.String("build_time", "20060102T150415"),
				slog.String("commit_hash", "deadbeef"),
			},
		},
	)

	//
	// Log at the INFO level.
	//
	slog.Info("hello world")

	// Want message string interpolation? Do this beforehand.
	variable := "there"
	slog.Info(fmt.Sprintf("hello %s", variable))

	// Log message with attributes
	slog.Info("hello with attributes", slog.Bool("simple", true), slog.Int("rating", 42))

	//
	// Log at the ERROR level.
	//
	err := errors.New("example")
	slog.Error("uh-oh", slog.Any("error", err))
}

// In a real application, [SetDefaults] probably only needs to be called once.
func ExampleSetDefaults() {
	//
	// Customize keys of metadata attributes.
	//
	logg.SetDefaults(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}),
		&logg.Settings{
			ApplicationMetadata: []slog.Attr{
				slog.String("branch_name", "main"),
				slog.String("build_time", "now"),
				slog.Any("foo", "bar"),
			},
			ApplicationMetadataKey: "application_metadata",
			TraceIDKey:             "request_uuid",
			DataKey:                "event_data",
		},
	)

	//
	// Output text format.
	//
	logg.SetDefaults(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}),
		&logg.Settings{ApplicationMetadata: []slog.Attr{slog.String("branch_name", "main")}},
	)

	//
	// Write to standard error as well as some file.
	//
	var file io.Writer
	sink := io.MultiWriter(os.Stderr, file)
	logg.SetDefaults(
		slog.NewJSONHandler(sink, &slog.HandlerOptions{Level: slog.LevelInfo}),
		&logg.Settings{ApplicationMetadata: []slog.Attr{slog.String("branch_name", "main")}},
	)
}

// Initialize the logger with or without data fields.
func ExampleNew() {
	// An empty input handler makes a logger that writes to the same handler
	// provisioned in SetupDefaults. This logger has its own data fields.
	logger := logg.New(nil, "", slog.Bool("whiskey", true), slog.Float64("tango", 1.23), slog.Int("foxtrot", 10))
	logger.Info("hello, world")

	// This logger also writes to the package configured handler, but doesn't
	// have any data fields of its own.
	loggerNoFields := logg.New(nil, "")
	loggerNoFields.Info("no data fields here")

	// This logger has its own tracing ID.
	loggerWithTraceID := logg.New(nil, "unique_trace_id")
	loggerWithTraceID.Info("that happened")

	// This logger writes to another Handler altogether, which might be useful
	// for testing.
	var handler slog.Handler
	loggerOwnHandler := logg.New(handler, "")
	loggerOwnHandler.Info("test")
}
