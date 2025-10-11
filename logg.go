package logg

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
)

var (
	root          *logger
	configureOnce sync.Once
	defaultSink   = os.Stderr
)

const loggLevelEnvVar = "LOGG_LEVEL"

// Configure initializes a root logger from which all subsequent logging events
// are derived, provided there are no previous writes to the log.  If there are
// any log writes before configuration, then all writes will go to os.Stderr by
// default. So, it's best to call this function as early as possible in your
// application.
//
// Configure will set up a prototype logger to write to w, include version
// metadata and may optionally write to moreSinks. The output destination(s)
// cannot be changed once this function is called.
//
// The version parameter may be empty, but it's recommended to put some metadata
// here so you can associate an event with the source code version.
func Configure(w io.Writer, version []slog.Attr, moreSinks ...io.Writer) {
	configureOnce.Do(func() {
		sinks := append([]io.Writer{w}, moreSinks...)
		m := io.MultiWriter(sinks...)
		lvl := os.Getenv(loggLevelEnvVar)
		lgr := newSlogger(m, lvl)
		root = newLogger(lgr, version, "")

		if strings.ToUpper(lvl) == "DEBUG" {
			root.lgr.Debug("configured logger")
		}
	})
}

// Errorf writes msg to the log at level error and additionally writes err to an
// error field. If msg is a format string and there are args, then it works like
// fmt.Printf.
func Errorf(err error, msg string, args ...interface{}) {
	r := rootLogger()
	m := fmt.Sprintf(msg, args...)
	log(context.Background(), r, slog.LevelError, err, m, "", r.attrs...)
}

// Infof writes msg to the log at level info. If msg is a format string and
// there are args, then it works like fmt.Printf.
func Infof(msg string, args ...interface{}) {
	r := rootLogger()
	m := fmt.Sprintf(msg, args...)
	log(context.Background(), r, slog.LevelInfo, nil, m, "", r.attrs...)
}

// An Emitter writes to the log at info or error levels.
type Emitter interface {
	Infof(msg string, args ...interface{})
	Errorf(err error, msg string, args ...interface{})
	WithID(ctx context.Context) Emitter
	WithData(attrs []slog.Attr) Emitter
}

func rootLogger() *logger {
	// fall back to default writer unless it's already configured.
	Configure(defaultSink, nil)

	return root
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

func log(ctx context.Context, l *logger, v slog.Level, err error, msg string, traceID string, attrs ...slog.Attr) {
	if !l.lgr.Enabled(ctx, v) {
		return
	}

	attrsToLog := make([]slog.Attr, 0, 4)
	if len(l.versioning) > 0 {
		attrsToLog = append(attrsToLog, slog.GroupAttrs(versionFieldName, l.versioning...))
	}

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

	l.lgr.LogAttrs(ctx, v, msg, attrsToLog...)
}
