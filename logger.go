package logg

import (
	"context"
	"log/slog"
)

// A logger emits events with preset versioning info and attrs data.
type logger struct {
	lgr        *slog.Logger
	versioning []slog.Attr
	id         string
	attrs      []slog.Attr
}

// New initializes a logger Emitter type and configures it so each event
// emission outputs attrs at the "data" key. If h is nil then it uses a
// default root handler, which is configured via the [Setup] function.
func New(h slog.Handler, dataAttrs ...slog.Attr) Emitter {
	defaults := rootLogger()
	if h == nil {
		h = defaults.lgr.Handler()
	}

	lgr := newSlogger(h, defaults.versioning...)
	out := newLogger(lgr, defaults.versioning, "", dataAttrs...)
	return out
}

func (l *logger) Debug(msg string, attrs ...slog.Attr) {
	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(context.Background(), l.lgr, slog.LevelDebug, nil, msg, l.id, mergedAttrs...)
}

func (l *logger) Info(msg string, attrs ...slog.Attr) {
	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(context.Background(), l.lgr, slog.LevelInfo, nil, msg, l.id, mergedAttrs...)
}

func (l *logger) Warn(msg string, attrs ...slog.Attr) {
	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(context.Background(), l.lgr, slog.LevelWarn, nil, msg, l.id, mergedAttrs...)
}

func (l *logger) Error(err error, msg string, attrs ...slog.Attr) {
	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(context.Background(), l.lgr, slog.LevelError, err, msg, l.id, mergedAttrs...)
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

func newLogger(lgr *slog.Logger, versioning []slog.Attr, id string, dataAttrs ...slog.Attr) *logger {
	return &logger{
		lgr:        lgr,
		versioning: versioning,
		id:         id,
		attrs:      dataAttrs,
	}
}

func newSlogger(handler slog.Handler, versioningAttrs ...slog.Attr) *slog.Logger {
	group := slog.GroupAttrs(versionFieldName, versioningAttrs...)
	lgr := slog.New(handler).With(group)
	lgr.Debug(libraryMsgPrefix + "initialized logger")
	return lgr
}
