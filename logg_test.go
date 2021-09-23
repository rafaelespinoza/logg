package logg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rafaelespinoza/logg"
)

func init() {
	logg.Configure(os.Stderr, map[string]string{"foo": "bar"})
}

func TestLogg(t *testing.T) {
	t.Run("Info", func(t *testing.T) {
		sink := newDataSink()
		logger := logg.New(map[string]interface{}{"sierra": "nevada"}, sink)

		// test logger
		logger.Infof("hello")
		testLogg(t, sink.Raw(), nil, "hello", false, map[string]interface{}{"sierra": "nevada"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// test event
		logger.WithData(map[string]interface{}{
			"bravo":   true,
			"delta":   234 * time.Millisecond,
			"foxtrot": float64(1.23),
			"india":   10,
		}).Infof("goodbye")
		testLogg(t, sink.Raw(), nil, "goodbye", false, map[string]interface{}{
			// The types of expected values are changed due to the way
			// encoding/json unmarshals interface{} values. This is far from
			// perfect, so at this point, considering the field tests to be
			// "close enough".
			"bravo":   true,
			"delta":   float64(234), // corresponding input is a time.Duration.
			"foxtrot": float64(1.23),
			"india":   float64(10), // corresponding input is an int.
			"sierra":  "nevada",
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// another event from same logger can have its own fields
		logger.WithData(map[string]interface{}{"zulu": true}).Infof("goodbye again")
		testLogg(t, sink.Raw(), nil, "goodbye again", false, map[string]interface{}{
			"zulu":   true,
			"sierra": "nevada",
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("Error", func(t *testing.T) {
		sink := newDataSink()
		logger := logg.New(map[string]interface{}{"sierra": "nevada"}, sink)

		// test logger
		logger.Errorf(errors.New("hello"), "logger error")
		testLogg(t, sink.Raw(), errors.New("hello"), "logger error", false, map[string]interface{}{"sierra": "nevada"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// test event
		logger.WithData(map[string]interface{}{
			"bravo":   false,
			"delta":   432 * time.Millisecond,
			"foxtrot": float64(3.21),
			"india":   100,
		}).Errorf(errors.New("goodbye"), "event error")
		testLogg(t, sink.Raw(), errors.New("goodbye"), "event error", false, map[string]interface{}{
			// The types of expected values are changed due to the way
			// encoding/json unmarshals interface{} values. This is far from
			// perfect, so at this point, considering the field tests to be
			// "close enough".
			"bravo":   false,
			"delta":   float64(432), // corresponding input is a time.Duration.
			"foxtrot": float64(3.21),
			"india":   float64(100), // corresponding input is an int.
			"sierra":  "nevada",
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// another event from same logger can have its own fields
		logger.WithData(map[string]interface{}{"zulu": true}).Errorf(errors.New("bye"), "goodbye again")
		testLogg(t, sink.Raw(), errors.New("bye"), "goodbye again", false, map[string]interface{}{
			"zulu":   true,
			"sierra": "nevada",
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestWithID(t *testing.T) {
	t.Run("logger passes ID to event", func(t *testing.T) {
		// The logger calls WithID, but the event does not.
		ctx := context.Background()
		sink := newDataSink()

		logger := logg.New(map[string]interface{}{"sierra": "nevada"}, sink).WithID(ctx)
		logger.Infof("logger with id")
		testLogg(t, sink.Raw(), nil, "logger with id", true, map[string]interface{}{"sierra": "nevada"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		logger.WithData(map[string]interface{}{"bravo": true}).Infof("event with id")
		testLogg(t, sink.Raw(), nil, "event with id", true, map[string]interface{}{"bravo": true, "sierra": "nevada"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("event can set its own ID", func(t *testing.T) {
		// The logger does not call WithID, but the event does.

		ctx := context.Background()
		sink := newDataSink()

		logger := logg.New(map[string]interface{}{"sierra": "nevada"}, sink)
		logger.Infof("logger without id")
		testLogg(t, sink.Raw(), nil, "logger without id", false, map[string]interface{}{"sierra": "nevada"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		logger.WithData(map[string]interface{}{"bravo": true}).WithID(ctx).Infof("event with own id")
		testLogg(t, sink.Raw(), nil, "event with own id", true, map[string]interface{}{
			"bravo":  true,
			"sierra": "nevada",
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestWithData(t *testing.T) {
	t.Run("allows input fields to replace existing fields", func(t *testing.T) {
		sink := newDataSink()

		logger := logg.New(map[string]interface{}{"foo": "alfa"}, sink)
		logger.Infof("a")
		testLogg(t, sink.Raw(), nil, "a", false, map[string]interface{}{"foo": "alfa"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		event := logger.WithData(map[string]interface{}{"foo": "bravo"})
		event.Infof("b")
		testLogg(t, sink.Raw(), nil, "b", false, map[string]interface{}{"foo": "bravo"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		otherEvent := event.WithData(map[string]interface{}{"foo": "charlie"})
		otherEvent.Infof("c")
		testLogg(t, sink.Raw(), nil, "c", false, map[string]interface{}{"foo": "charlie"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// check that first logger's original data hasn't changed unexpectedly.
		logger.Infof("d")
		testLogg(t, sink.Raw(), nil, "d", false, map[string]interface{}{"foo": "alfa"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// check that first event's original data hasn't changed unexpectedly.
		event.Infof("e")
		testLogg(t, sink.Raw(), nil, "e", false, map[string]interface{}{"foo": "bravo"})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func testLogg(t *testing.T, in []byte, expErr error, expMessage string, expTraceID bool, expData map[string]interface{}) {
	t.Helper()

	const (
		errorKey   = "error"
		messageKey = "message"
		traceIDKey = "x_trace_id"
		dataKey    = "data"
	)

	var (
		parsedRoot map[string]interface{}
		parsedData map[string]interface{}
	)

	if err := json.Unmarshal(in, &parsedRoot); err != nil {
		t.Fatal(err)
	}

	// test error
	if expErr != nil {
		if val, ok := parsedRoot[errorKey]; !ok {
			t.Errorf("expected to have key %q", messageKey)
		} else if val.(string) != expErr.Error() {
			t.Errorf("wrong value at %q; got %q, expected %q", errorKey, val.(string), expErr.Error())
		}
	} else {
		val, ok := parsedRoot[errorKey]
		if ok {
			t.Errorf("unexpected error at %q; got %v", errorKey, val)
		}
	}

	// test message
	if val, ok := parsedRoot[messageKey]; !ok {
		t.Errorf("expected to have key %q", messageKey)
	} else if val.(string) != expMessage {
		t.Errorf("wrong value at %q; got %q, expected %q", messageKey, val.(string), expMessage)
	}

	// test trace id
	numTraceKeyValues := strings.Count(string(in), traceIDKey)
	if expTraceID {
		if val, ok := parsedRoot[traceIDKey]; !ok {
			t.Errorf("expected to have key %q", traceIDKey)
		} else if val.(string) == "" {
			t.Errorf("expected non-empty value at %q", traceIDKey)
		}
		if numTraceKeyValues != 1 {
			t.Errorf("wrong count of %q values; got %d, expected %d", traceIDKey, numTraceKeyValues, 1)
		}
	} else {
		val, ok := parsedRoot[traceIDKey]
		if ok {
			t.Errorf("unexpected trace ID at %q; got %v", traceIDKey, val)
		}
		if numTraceKeyValues > 0 {
			t.Errorf("wrong count of %q values; got %d, expected %d", traceIDKey, numTraceKeyValues, 0)
		}
	}

	// test data
	if val, ok := parsedRoot[dataKey]; !ok {
		t.Errorf("expected to have key %q", dataKey)
	} else if parsedData, ok = val.(map[string]interface{}); !ok {
		t.Errorf("expected %q to be a %T", dataKey, make(map[string]interface{}))
	}

	if len(parsedData) != len(expData) {
		t.Errorf("wrong number of keys; got %d, expected %d", len(parsedData), len(expData))
	}

	for expKey, expVal := range expData {
		got, ok := parsedData[expKey]
		if !ok {
			t.Errorf("expected to have subkey [%q][%q]", dataKey, expKey)
		} else if got != expVal {
			t.Errorf(
				"wrong value at [%q][%q]; got %v (type %T), expected %v (type %T)",
				dataKey, expKey, got, got, expVal, expVal,
			)
		}
	}
}

func newDataSink() *DataSink {
	var buf bytes.Buffer
	return &DataSink{buf: &buf}
}

// DataSink is designed to capture one logging entry at a time in a buffer.
type DataSink struct{ buf *bytes.Buffer }

// Write resets the internal buffer and replaces it with a logging entry.
func (s *DataSink) Write(in []byte) (n int, e error) {
	s.buf.Reset()
	n, e = s.buf.Write(in)
	return
}

// Raw outputs the buffer contents for inspection.
func (s *DataSink) Raw() []byte { return s.buf.Bytes() }
