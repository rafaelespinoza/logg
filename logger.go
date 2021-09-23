package logg

import (
	"context"
	"io"

	"github.com/rs/zerolog"
)

// A logger emits events with preset fields.
type logger struct {
	context *zerolog.Context
	fields  map[string]interface{}
}

// New initializes a logger Emitter type and configures it so each event
// emission outputs fields at the data key. If sinks is empty or the first sink
// is nil, then it writes to the same destination as the root logger. If sinks
// is non-empty then it duplicates the root logger and writes to sinks.
func New(fields map[string]interface{}, sinks ...io.Writer) Emitter {
	var sub zerolog.Context
	if len(sinks) == 0 || sinks[0] == nil {
		sub = rootLogger().With()
	} else {
		m := zerolog.MultiLevelWriter(sinks...)
		sub = rootLogger().Output(m).With()
	}

	return &logger{context: &sub, fields: shallowDupe(fields)}
}

func (l *logger) Errorf(err error, msg string, args ...interface{}) {
	lgr := l.context.Logger()
	newZerologErrorEvent(&lgr, err, l.fields).Msgf(msg, args...)
}

func (l *logger) Infof(msg string, args ...interface{}) {
	lgr := l.context.Logger()
	newZerologInfoEvent(&lgr, l.fields).Msgf(msg, args...)
}

func (l *logger) WithID(ctx context.Context) Emitter {
	lgr := l.context.Logger()
	l.context = newZerologCtxWithID(ctx, &lgr)
	return l
}

// WithData prepares a logging entry and captures any event-specific data in
// fields. Call the Emitter methods to write to the log.
func (l *logger) WithData(fields map[string]interface{}) Emitter {
	logger := l.context.Logger()

	// use original l.fields as a base, but let the input fields override any
	// conflict keys for the output event.
	tmp := shallowDupe(l.fields)
	dupedFields := mergeFields(tmp, fields)

	return &event{
		logger: &logger,
		fields: dupedFields,
	}
}
