package logg

import (
	"context"
	"log/slog"
)

// An Emitter logs events with preset version info and attributes data. Set a
// tracing ID on a context.Context with the [SetID] function and pass it into
// the [Emitter.WithID] method. Log output methods are Debug, Info, Warn, Error
// which write at the respective logging levels.
//
// # Data attribute management
//
// This section describes how attributes are managed for the output log's "data"
// key. When an Emitter is created via the [New] function it takes slog.Attr as
// input and prepares the attributes to be present in future logging events for
// the created Emitter. The [Emitter.WithData] method combines the input
// attributes with the Emitter's existing attributes and returns a new Emitter
// with the resulting attributes. When there is an attribute with the same Key
// in either list of attributes, no deduplication or merging is performed. This
// is because the slog.Handler may already deduplicate by default, may want to
// conditionally deduplicate the attributes, or may be designed to output those
// attributes anyways.
//
// The log output methods, when passed input attributes and when the Emitter is
// enabled, also apply this logic. A notable difference with the
// [Emitter.WithData] method is that the created attributes are not set to the
// Emitter. Instead, they are sent directly to the log.
type Emitter struct {
	lgr        *slog.Logger
	versioning []slog.Attr
	id         string
	attrs      []slog.Attr
}

// New initializes a logger type and configures it so each event emission
// outputs attrs at the "data" key. If h is nil then it uses a default root
// handler, which is configured via the [Setup] function.
func New(h slog.Handler, dataAttrs ...slog.Attr) *Emitter {
	defaults := rootEmitter()
	if h == nil {
		h = defaults.lgr.Handler()
	}

	lgr := newSlogger(h, defaults.versioning...)
	out := newEmitter(lgr, defaults.versioning, "", dataAttrs...)
	return out
}

// Debug writes msg at the DEBUG level with optional attributes.
func (l *Emitter) Debug(msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelDebug
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := combineAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, nil, msg, l.id, mergedAttrs...)
}

// Info writes msg at the INFO level with optional attributes.
func (l *Emitter) Info(msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelInfo
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := combineAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, nil, msg, l.id, mergedAttrs...)
}

// Warn writes msg at the WARN level with optional attributes.
func (l *Emitter) Warn(msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelWarn
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := combineAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, nil, msg, l.id, mergedAttrs...)
}

// Error writes msg and err to the log at the ERROR level with optional
// attributes. The input err is placed on to the "error" key in the log event.
func (l *Emitter) Error(err error, msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelError
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := combineAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, err, msg, l.id, mergedAttrs...)
}

// WithID reads the input ctx and prepares the Emitter to write the tracing ID
// at the key, "x_trace_id". Use the [SetID] function to prepare the input
// context. If a tracing ID already existed on the Emitter, then it's replaced.
func (l *Emitter) WithID(ctx context.Context) *Emitter {
	l.id, _ = GetID(ctx)
	return l
}

// WithData creates a new Emitter with added attributes. When there is an
// existing attribute with the same key as one of the input attributes, then
// both are accepted. How this data is presented in the log output depends on
// what the slog.Handler does for attributes.
func (l *Emitter) WithData(attrs []slog.Attr) *Emitter {
	versioning := rootEmitter().versioning
	mergedAttrs := combineAttrs(l.attrs, attrs)

	return newEmitter(l.lgr, versioning, l.id, mergedAttrs...)
}

func newEmitter(lgr *slog.Logger, versioning []slog.Attr, id string, dataAttrs ...slog.Attr) *Emitter {
	return &Emitter{
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

// combineAttrs combines the input lists into an output list. It returns early
// if 1 or both lists are length 0.
func combineAttrs(prevList, nextList []slog.Attr) []slog.Attr {
	if len(prevList) > 0 && len(nextList) < 1 {
		return prevList
	} else if len(prevList) < 1 && len(nextList) > 0 {
		return nextList
	} else if len(prevList) < 1 && len(nextList) < 1 {
		return make([]slog.Attr, 0)
	}

	out := make([]slog.Attr, len(prevList), len(prevList)+len(nextList))
	copy(out, prevList)
	return append(out, nextList...)
}
