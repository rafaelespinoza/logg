package logg

import (
	"context"

	"github.com/rs/zerolog"
)

type event struct {
	logger *zerolog.Logger
	fields map[string]interface{}
}

func (e *event) Infof(msg string, args ...interface{}) {
	newZerologInfoEvent(e.logger, e.fields).Msgf(msg, args...)
}

func (e *event) Errorf(err error, msg string, args ...interface{}) {
	newZerologErrorEvent(e.logger, err, e.fields).Msgf(msg, args...)
}

// WithID sets a tracing ID on the logging entry. If the event is constructed
// with a Logger, which has already called WithID, then calling this method will
// add another trace ID key-value pair at the top of the logging entry. This
// behavior is documented in the logging library, github.com/rs/zerolog README.
//
// There are some known techniques for working with this limitation. One is to
// not use this method if you know that the event has been constructed with a
// logger that already has an event ID. The logger's trace ID will be available
// on this event's context. Another is to use the ID function to create an ID on
// a context.Context and use the same output context on the logger and the
// event. This will duplicate the key, but the trace ID values will be the same.
func (e *event) WithID(ctx context.Context) Emitter {
	// I've attempted to find ways to exclude a key from the logger context
	// while replacing it with one produced here, but the logging library does
	// not have an API to overwrite or replace those existing values, nor does
	// it have a way to write to the same io.Writer destination without copying
	// all the fields.
	lgr := newZerologCtxWithID(ctx, e.logger).Logger()
	e.logger = &lgr
	return e
}

const eventDebugWithDataMsg = "called WithData on an event; prefer calling WithData on a logger type"

func (e *event) WithData(fields map[string]interface{}) Emitter {
	// I'd think this method would only be used seldomly. This method replaces e
	// in order to accept and merge fields into e.fields, preferring the new
	// data in fields over any potentially conflicting keys in e.fields, but
	// also try not to change to the original e.fields.
	rootLogger().Debug().Msg(eventDebugWithDataMsg)

	tmp := shallowDupe(e.fields)
	dupedFields := mergeFields(tmp, fields)

	return &event{
		logger: e.logger,
		fields: dupedFields,
	}
}
