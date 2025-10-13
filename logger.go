package logg

import (
	"cmp"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// A logger emits events with preset versioning info and attrs data.
type logger struct {
	lgr        *slog.Logger
	versioning []slog.Attr
	id         string
	attrs      []slog.Attr
}

// New initializes a logger Emitter type and configures it so each event
// emission outputs attrs at the data key. If sinks is empty or the first sink
// is nil, then it writes to the same destination as the root logger. If sinks
// is non-empty then it duplicates the root logger and writes to sinks.
func New(attrs []slog.Attr, sinks ...io.Writer) Emitter {
	if len(sinks) < 1 || sinks[0] == nil {
		sinks = []io.Writer{defaultSink}
	}
	m := io.MultiWriter(sinks...)
	lvl := cmp.Or(os.Getenv(loggLevelEnvVar), slog.LevelInfo.String())
	lgr := newSlogger(m, lvl)
	return newLogger(lgr, rootLogger().versioning, "", attrs...)
}

func (l *logger) Error(err error, msg string, attrs ...slog.Attr) {
	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(context.Background(), l, slog.LevelError, err, msg, l.id, mergedAttrs...)
}

func (l *logger) Info(msg string, attrs ...slog.Attr) {
	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(context.Background(), l, slog.LevelInfo, nil, msg, l.id, mergedAttrs...)
}

func (l *logger) WithID(ctx context.Context) Emitter {
	l.id, _ = GetID(ctx)
	return l
}

// WithData prepares a logging entry and captures any event-specific data in
// attrs. Call the Emitter methods to write to the log.
func (l *logger) WithData(attrs []slog.Attr) Emitter {
	versioning := rootLogger().versioning

	// use original l.attrs as a base, but let the input attrs override any
	// conflict keys for the output event.
	mergedAttrs := mergeAttrs(l.attrs, attrs)

	return newLogger(l.lgr, versioning, l.id, mergedAttrs...)
}

func newLogger(lgr *slog.Logger, versioning []slog.Attr, id string, attrs ...slog.Attr) *logger {
	out := logger{
		lgr:        lgr,
		versioning: versioning,
		id:         id,
		attrs:      attrs,
	}

	return &out
}

func newSlogger(w io.Writer, level string) *slog.Logger {
	var lvl slog.Level
	level = strings.ToUpper(strings.TrimSpace(level))
	switch level {
	case slog.LevelDebug.String():
		lvl = slog.LevelDebug
	case slog.LevelInfo.String():
		lvl = slog.LevelInfo
	case slog.LevelWarn.String():
		lvl = slog.LevelWarn
	case slog.LevelError.String():
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
		slog.Warn("unknown level, setting default", "unknown_level", level, "default", lvl.String())
	}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl})

	lgr := slog.New(handler)
	lgr.Debug("initialized logger")
	return lgr
}
