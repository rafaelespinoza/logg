package logg_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rafaelespinoza/logg"
)

func TestEvent(t *testing.T) {
	const (
		traceIDKey = "x_trace_id"
		traceIDVal = "event_id"
	)
	var (
		parsedRoot        map[string]interface{}
		numTraceKeyValues int
	)

	ctx := logg.SetID(context.Background(), traceIDVal)
	sink := newDataSink()
	logger := logg.New(map[string]interface{}{"sierra": "nevada"}, sink).WithID(ctx)

	logger.WithData(map[string]interface{}{"bravo": true}).Infof("test")
	if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
		t.Fatal(err)
	}
	if id, ok := parsedRoot[traceIDKey]; !ok {
		t.Errorf("expected output to have key %q", traceIDKey)
	} else if id.(string) == "" {
		t.Errorf("expected non-empty value at %q", traceIDKey)
	}

	numTraceKeyValues = strings.Count(string(sink.Raw()), traceIDKey)
	if numTraceKeyValues != 1 {
		t.Errorf("wrong count of %q values; got %d, expected %d", traceIDKey, numTraceKeyValues, 1)
	}
	if t.Failed() {
		t.Logf("%s", sink.Raw())
	}

	// Same logger context on a new event yields the same ID
	logger.WithData(map[string]interface{}{"bravo": true}).Infof("test")
	if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
		t.Fatal(err)
	}
	if id, ok := parsedRoot[traceIDKey]; !ok {
		t.Errorf("expected output to have key %q", traceIDKey)
	} else if id.(string) != traceIDVal {
		t.Errorf("wrong id; got %q, expected %q", id.(string), traceIDVal)
	}
	numTraceKeyValues = strings.Count(string(sink.Raw()), traceIDKey)
	if numTraceKeyValues != 1 {
		t.Errorf("wrong count of %q values; got %d, expected %d", traceIDKey, numTraceKeyValues, 1)
	}
	if t.Failed() {
		t.Logf("%s", sink.Raw())
	}

	// When the logger context has an ID, is passed to the event but the event
	// also calls WithID with another context, it replaces the ID.
	const differentTraceID = "different_" + traceIDVal
	ctxB := logg.SetID(ctx, differentTraceID)

	logger.WithData(map[string]interface{}{"bravo": true}).WithID(ctxB).Infof("test")
	if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
		t.Fatal(err)
	}
	if id, ok := parsedRoot[traceIDKey]; !ok {
		t.Errorf("expected output to have key %q", traceIDKey)
	} else if id.(string) != differentTraceID {
		t.Errorf("wrong id; got %q, expected %q", id.(string), differentTraceID)
	}
	// Documents known behavior stemming from the logging library, which doesn't
	// do any field de-duplication. Calling WithID multiple times just adds
	// another id. See the github.com/rs/zerolog README for more info.
	numTraceKeyValues = strings.Count(string(sink.Raw()), traceIDKey)
	if numTraceKeyValues != 1 {
		t.Logf(
			"info about number of %q values; got %d, but %d would be ideal; %s",
			traceIDKey, numTraceKeyValues, 1, `¯\_(ツ)_/¯`,
		)
	}
	if t.Failed() {
		t.Logf("%s", sink.Raw())
	}
}
