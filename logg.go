package logg

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sort"
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
func Configure(w io.Writer, version map[string]string, moreSinks ...io.Writer) {
	configureOnce.Do(func() {
		sinks := append([]io.Writer{w}, moreSinks...)
		m := io.MultiWriter(sinks...)
		lvl := os.Getenv(loggLevelEnvVar)

		versioningData := []slog.Attr{}
		if version != nil {
			versioningData = make([]slog.Attr, len(version))
			for i, key := range slices.Sorted(maps.Keys(version)) {
				versioningData[i] = slog.String(key, version[key])
			}
		}

		lgr := newSlogger(m, lvl)
		root = newLogger(lgr, versioningData, "", make(map[string]any))

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
	log(context.Background(), r, slog.LevelError, err, m, "", r.fields)
}

// Infof writes msg to the log at level info. If msg is a format string and
// there are args, then it works like fmt.Printf.
func Infof(msg string, args ...interface{}) {
	r := rootLogger()
	m := fmt.Sprintf(msg, args...)
	log(context.Background(), r, slog.LevelInfo, nil, m, "", r.fields)
}

// An Emitter writes to the log at info or error levels.
type Emitter interface {
	Infof(msg string, args ...interface{})
	Errorf(err error, msg string, args ...interface{})
	WithID(ctx context.Context) Emitter
	WithData(fields map[string]interface{}) Emitter
}

func rootLogger() *logger {
	// fall back to default writer unless it's already configured.
	Configure(defaultSink, nil)

	return root
}

func shallowDupe(in map[string]interface{}) (out map[string]interface{}) {
	out = make(map[string]interface{})
	if in == nil {
		return
	}
	for key, val := range in {
		out[key] = val
	}
	return
}

func mergeFields(dst, src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return dst
	}
	for key, val := range src {
		dst[key] = val
	}
	return dst
}

func mapToAttrs(fields map[string]interface{}) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(fields))

	for key, val := range fields {
		attrs = append(attrs, slog.Attr{Key: key, Value: slog.AnyValue(val)})
	}

	sort.Slice(attrs, func(i, j int) bool { return attrs[i].Key < attrs[j].Key })
	return attrs
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

func log(ctx context.Context, l *logger, v slog.Level, err error, msg string, traceID string, fields map[string]interface{}) {
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

	if len(fields) > 0 {
		attrsToLog = append(attrsToLog, slog.GroupAttrs(dataFieldName, mapToAttrs(fields)...))
	}

	attrsToLog = slices.Clip(attrsToLog)

	l.lgr.LogAttrs(ctx, v, msg, attrsToLog...)
}
