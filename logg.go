package logg

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	root          zerolog.Context
	configureOnce sync.Once
	defaultSink   = os.Stderr
)

// dataFieldName is the logging entry key for any event-specific data.
const dataFieldName = "data"

// Configure initializes a root logger from with all subsequent logging events
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
		m := zerolog.MultiLevelWriter(sinks...)
		root = zerolog.New(m).With().Timestamp()

		if version != nil {
			dict := zerolog.Dict()
			for key, val := range version {
				dict = dict.Str(key, val)
			}
			root = root.Dict("version", dict)
		}

		lgr := root.Logger()
		lgr.Info().Msg("configured logger")
	})
}

// Errorf writes msg to the log at level error and additionally writes err to an
// error field. If msg is a format string and there are args, then it works like
// fmt.Printf.
func Errorf(err error, msg string, args ...interface{}) {
	rootLogger().Err(err).Msgf(msg, args...)
}

// Infof writes msg to the log at level info. If msg is a format string and
// there are args, then it works like fmt.Printf.
func Infof(msg string, args ...interface{}) {
	rootLogger().Info().Msgf(msg, args...)
}

// An Emitter emitter writes to the log at info or error levels.
type Emitter interface {
	Infof(msg string, args ...interface{})
	Errorf(err error, msg string, args ...interface{})
	WithID(ctx context.Context) Emitter
	WithData(fields map[string]interface{}) Emitter
}

func rootLogger() *zerolog.Logger {
	// fall back to default writer unless it's already configured.
	Configure(defaultSink, nil)

	out := root.Logger()
	return &out
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

func newZerologInfoEvent(lgr *zerolog.Logger, fields map[string]interface{}) *zerolog.Event {
	return lgr.Info().Dict(dataFieldName, zerolog.Dict().Fields(fields))
}

func newZerologErrorEvent(lgr *zerolog.Logger, err error, fields map[string]interface{}) *zerolog.Event {
	return lgr.Err(err).Dict(dataFieldName, zerolog.Dict().Fields(fields))
}
