package logg

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"sync/atomic"
)

var root atomic.Pointer[Emitter]

const libraryMsgPrefix = "logg: "

func init() {
	defaultOutput := os.Stderr
	defaultLevel := slog.LevelInfo
	defaultHandler := slog.NewJSONHandler(defaultOutput, &slog.HandlerOptions{Level: defaultLevel})
	Setup(defaultHandler)
}

// Setup initializes a package-level prototype logger from which subsequent logs
// are based upon. The default settings are to write to os.Stderr at the INFO
// level in line-delimited JSON format. Top-level functions such as Info and
// Error use this logger. Loggers created with [New] consider values set in this
// function as defaults.
//
// The version parameter may be empty, but it's recommended to put some metadata
// here so you can associate an event with the source code version. The
// attributes in version will be part of "version" group for subsequent log
// events, including those output by an [Emitter].
func Setup(h slog.Handler, version ...slog.Attr) {
	lgr := newSlogger(h, version...)
	r := newEmitter(lgr, version, "")

	root.Store(r)
	rootEmitter().lgr.Debug(libraryMsgPrefix + "setup root logger")
}

// Debug writes msg to the package logger with optional attributes at level
// DEBUG.
func Debug(msg string, attrs ...slog.Attr) {
	log(context.Background(), rootEmitter().lgr, slog.LevelDebug, nil, msg, "", attrs...)
}

// Info writes msg to the package logger with optional attributes at level INFO.
func Info(msg string, attrs ...slog.Attr) {
	log(context.Background(), rootEmitter().lgr, slog.LevelInfo, nil, msg, "", attrs...)
}

// Warn writes msg to the package logger with optional attributes at level WARN.
func Warn(msg string, attrs ...slog.Attr) {
	log(context.Background(), rootEmitter().lgr, slog.LevelWarn, nil, msg, "", attrs...)
}

// Error writes msg to the package logger with optional attributes at level
// ERROR and additionally writes err to the output's "error" field.
func Error(err error, msg string, attrs ...slog.Attr) {
	log(context.Background(), rootEmitter().lgr, slog.LevelError, err, msg, "", attrs...)
}

func rootEmitter() *Emitter {
	return root.Load()
}

// Logging entry keys.
const (
	// versionFieldName is the logging entry key for application versioning info.
	versionFieldName = "version"
	// errorFieldName is the logging entry key for an error.
	errorFieldName = "error"
	// traceIDFieldName is the logging entry key for a tracing ID.
	traceIDFieldName = "x_trace_id"
	// dataFieldName is the logging entry key for any event-specific data.
	dataFieldName = "data"
)

func log(ctx context.Context, lgr *slog.Logger, v slog.Level, err error, msg string, traceID string, attrs ...slog.Attr) {
	if !lgr.Enabled(ctx, v) {
		return
	}

	attrsToLog := make([]slog.Attr, 0, 3)

	if err != nil {
		attrsToLog = append(attrsToLog, slog.Any(errorFieldName, err))
	}

	if traceID != "" {
		attrsToLog = append(attrsToLog, slog.String(traceIDFieldName, traceID))
	}

	if len(attrs) > 0 {
		attrsToLog = append(attrsToLog, slog.GroupAttrs(dataFieldName, attrs...))
	}

	attrsToLog = slices.Clip(attrsToLog)

	lgr.LogAttrs(ctx, v, msg, attrsToLog...)
}
