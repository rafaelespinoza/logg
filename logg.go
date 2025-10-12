package logg

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"sync/atomic"
)

var root atomic.Pointer[logger]

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
// events, including those an Emitter created via [New].
func Setup(h slog.Handler, version ...slog.Attr) {
	lgr := newSlogger(h, version...)
	r := newLogger(lgr, version, "")

	root.Store(r)
	rootLogger().lgr.Debug(libraryMsgPrefix + "setup root logger")
}

// Error writes msg to the log at level ERROR and additionally writes err to an
// error field.
func Error(err error, msg string, attrs ...slog.Attr) {
	log(context.Background(), rootLogger().lgr, slog.LevelError, err, msg, "", attrs...)
}

// Info writes msg to the log at level INFO.
func Info(msg string, attrs ...slog.Attr) {
	log(context.Background(), rootLogger().lgr, slog.LevelInfo, nil, msg, "", attrs...)
}

// An Emitter writes to the log at info or error levels.
type Emitter interface {
	Info(msg string, attrs ...slog.Attr)
	Error(err error, msg string, attrs ...slog.Attr)
	WithID(ctx context.Context) Emitter
	WithData(attrs []slog.Attr) Emitter
}

func rootLogger() *logger {
	return root.Load()
}

// mergeAttrs combines and deduplicates the input lists into a new list
// according to this logic:
//
//   - if a key is duplicated within the same list, then the item that appears
//     later in that list has higher precedence.
//   - if a key is found in both lists, then the item in nextList has higher
//     precedence than an item with the same key from prevList.
//   - if a key is unique to either list, then it will be in the output.
//
// The output list is a new slice, neither input list is modified.
func mergeAttrs(prevList, nextList []slog.Attr) []slog.Attr {
	maxLength := max(len(prevList), len(nextList))

	// 2 passes over each list.
	// 1st pass: collect items to add.
	// 2nd pass: build output while preserving original order of keys.

	//
	// 1st pass
	//
	attrsBykey := make(map[string]slog.Attr, maxLength)

	// Start with items from prevList. Items that are later in prevList take
	// precedence over items with the same key that appear earlier in prevList.
	for _, attr := range prevList {
		attrsBykey[attr.Key] = attr
	}

	// Then look through nextList. Same idea as prevList: allow later items with
	// same key to take precedence over earlier items. But also, any item in
	// nextList with the same key as an item in prevList will take precedence.
	for _, attr := range nextList {
		attrsBykey[attr.Key] = attr
	}

	//
	// 2nd pass
	//
	addedKeys := make(map[string]struct{}, maxLength)
	out := make([]slog.Attr, 0, maxLength)
	var found bool

	// Look at items collected from nextList, and add them to output unless
	// we've already seen the value.
	for _, attr := range nextList {
		if _, found = addedKeys[attr.Key]; !found {
			out = append(out, attrsBykey[attr.Key])
			addedKeys[attr.Key] = struct{}{} // prevent further re-adding to output list
		}
	}

	// Now look at items from prevList, and only add the item if its key is
	// unique to prevList.
	for _, attr := range prevList {
		if _, found = addedKeys[attr.Key]; !found {
			out = append(out, attrsBykey[attr.Key])
			addedKeys[attr.Key] = struct{}{} // prevent further re-adding to output list
		}
	}

	return slices.Clip(out)
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
