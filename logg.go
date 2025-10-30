// Package logg is a thin wrapper around log/slog to decorate a [slog.Logger]
// with commonly-needed metadata for applications and events. Events will have
// the same keys described by the slog package for built-in attributes; ie:
// "time", "level", "msg", "source".
//
// Additionally, these default top-level keys may be present:
//
//   - "application_metadata": []slog.Attr. ie: version control data, build
//     times, etc.
//   - "trace_id": string. A tracing ID, may be present for output from
//     loggers created via [New]
//   - "data": []slog.Attr. Event-specific attributes, may be present for output
//     from loggers created with [New].
//
// These key names may be configured with the [SetDefaults] function.
package logg

import (
	"cmp"
	"log/slog"
	"slices"

	"github.com/rafaelespinoza/logg/internal"
)

const libraryMsgPrefix = "logg: "

// Settings is package configuration data for overriding package defaults via
// [SetDefaults]. ApplicationMetadata could be versioning metadata, ie: commit
// hash, build time, etc. The placement of this data within a logging event can
// be set with ApplicationMetadataKey. The location of event tracing IDs can be
// set with TraceIDKey. Other event-specific attributes would be placed at
// DataKey.
type Settings struct {
	ApplicationMetadata    []slog.Attr
	ApplicationMetadataKey string
	TraceIDKey             string
	DataKey                string
}

var defaults = Settings{
	ApplicationMetadata:    []slog.Attr{},
	ApplicationMetadataKey: "application_metadata",
	TraceIDKey:             "trace_id",
	DataKey:                "data",
}

// defaultHandler is an untouched slog.Handler meant as a reference for building
// new loggers. Calling the .WithAttrs or .WithGroup methods on a [slog.Handler]
// returns a new Handler with different attributes or groups than before. When
// building new loggers in this package, we want to start with something that
// doesn't already have any groups or attributes added to it yet. Therefore, we
// can't use the Handler for [slog.Default]; at least not after calling the
// [slog.SetDefault] function later on from [SetDefaults]. For the
// initialization of this package, trust that initial value retrieved from slog
// is also an untouched handler.
var defaultHandler = slog.Default().Handler()

// SetDefaults initializes a package-level prototype logger from which
// subsequent logs are based upon. This function is intended to be called only
// once, shortly after application startup. It will also call the
// [slog.SetDefault] function, thereby affecting all output functionality in
// that package. Loggers created with [New] consider values set in this function
// as defaults. When the handler input h, is empty, then the current Handler
// value of [slog.Default] is used. When the input settings is empty, then
// defaults are used. Default settings are mentioned in the package
// documentation.
//
// The side effects of this function may put this data at top-level keys for
// every log entry:
//   - application metadata.
//   - a trace ID value, only present when a logger is created via [New].
//   - event-specific  data attributes.
func SetDefaults(h slog.Handler, settings *Settings) {
	if settings == nil {
		settings = &Settings{}
	}

	appMetadata := defaults.ApplicationMetadata
	// either you want to set the value to something, or you want to unset it
	// altogether.
	if len(settings.ApplicationMetadata) > 0 || settings.ApplicationMetadata == nil {
		appMetadata = settings.ApplicationMetadata
	}
	if h == nil {
		h = defaultHandler
	}

	nextSettings := Settings{
		ApplicationMetadata:    slices.Clone(appMetadata),
		ApplicationMetadataKey: cmp.Or(settings.ApplicationMetadataKey, defaults.ApplicationMetadataKey),
		TraceIDKey:             cmp.Or(settings.TraceIDKey, defaults.TraceIDKey),
		DataKey:                cmp.Or(settings.DataKey, defaults.DataKey),
	}
	defaults = nextSettings
	defaultHandler = h

	nextSlogger := newSlogger(h, defaults, "")
	slog.SetDefault(nextSlogger)

	slog.Debug(libraryMsgPrefix + "set default logger")
}

// New creates a [slog.Logger], with the same configuration and handler set via
// [SetDefaults]. The traceID, if non-empty, will be at a top-level key for
// subsequent logging output with the created Logger. Similarly, dataAttrs will
// be grouped at a top-level key.
func New(traceID string, dataAttrs ...slog.Attr) *slog.Logger {
	h := defaultHandler
	data := attrsToAnys(dataAttrs...)
	return newSlogger(h, defaults, traceID).
		WithGroup(defaults.DataKey).With(data...)
}

func newSlogger(handler slog.Handler, settings Settings, traceID string) *slog.Logger {
	attrs := make([]slog.Attr, 0, 2)
	attrs = append(attrs, internal.SlogGroupAttrs(settings.ApplicationMetadataKey, settings.ApplicationMetadata...))
	if traceID != "" {
		attrs = append(attrs, slog.String(settings.TraceIDKey, traceID))
	}
	args := attrsToAnys(attrs...)

	lgr := slog.New(handler).With(args...)
	lgr.Debug(libraryMsgPrefix + "initialized logger")
	return lgr
}

func attrsToAnys(in ...slog.Attr) (out []any) {
	out = make([]any, len(in))
	for i, attr := range in {
		out[i] = attr
	}
	return
}
