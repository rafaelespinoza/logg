package logg_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rafaelespinoza/logg"
)

func init() {
	setupPackageVars()
}

// setupPackageVars sets up some package-level state expected by most tests.
// Output for package-level logging functions will be written to w.
func setupPackageVars() {
	// appMetadata should be present in all subsequent log entries; not only
	// from package-level functions, but also from Logger instances.
	appMetadata := []slog.Attr{slog.String("branch_name", "dev"), slog.String("build_time", "now")}

	// Tests rely on parsing the JSON log entries to check for correctness.
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})

	// Restore defaults
	logg.SetDefaults(handler, &logg.Settings{
		ApplicationMetadata:    appMetadata,
		ApplicationMetadataKey: "application_metadata",
		TraceIDKey:             "trace_id",
		DataKey:                "data",
	})
}

func TestSetDefaults(t *testing.T) {
	t.Cleanup(func() { setupPackageVars() })

	t.Run("handler level", func(t *testing.T) {
		tests := []struct {
			inputLevel     slog.Level
			expDebugOutput bool
			expInfoOutput  bool
			expWarnOutput  bool
			expErrorOutput bool
		}{
			{
				inputLevel:     slog.LevelDebug,
				expDebugOutput: true, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelInfo,
				expDebugOutput: false, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelWarn,
				expDebugOutput: false, expInfoOutput: false, expWarnOutput: true, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelError,
				expDebugOutput: false, expInfoOutput: false, expWarnOutput: false, expErrorOutput: true,
			},
			{
				inputLevel:     slog.LevelError + 1,
				expDebugOutput: false, expInfoOutput: false, expWarnOutput: false, expErrorOutput: false,
			},
		}

		for _, test := range tests {
			t.Run(test.inputLevel.String(), func(t *testing.T) {
				t.Cleanup(func() { setupPackageVars() })

				sink, handler := newDataSinkAndJSONHandler(test.inputLevel)
				logg.SetDefaults(handler, nil)

				slog.Debug(t.Name())
				gotDebug := sink.Raw()
				if test.expDebugOutput && len(gotDebug) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelDebug.String())
				} else if !test.expInfoOutput && len(gotDebug) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelDebug.String())
				}

				slog.Info(t.Name())
				gotInfo := sink.Raw()
				if test.expInfoOutput && len(gotInfo) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelInfo.String())
				} else if !test.expInfoOutput && len(gotInfo) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelInfo.String())
				}

				slog.Warn(t.Name())
				gotWarn := sink.Raw()
				if test.expWarnOutput && len(gotWarn) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelWarn.String())
				} else if !test.expWarnOutput && len(gotWarn) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelWarn.String())
				}

				slog.Error(t.Name(), slog.Any("error", errors.New("test")))
				gotError := sink.Raw()
				if test.expErrorOutput && len(gotError) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelError.String())
				} else if !test.expErrorOutput && len(gotError) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelError.String())
				}
			})
		}
	})

	t.Run("handler format TEXT", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink := newDataSink()
		handler := slog.NewTextHandler(sink, &slog.HandlerOptions{Level: slog.LevelInfo})
		logg.SetDefaults(handler, nil)

		slog.Info(t.Name())

		logEntry := sink.Raw()
		if !strings.Contains(string(logEntry), "TEXT") {
			t.Fatalf("expected output %q to contain %q", logEntry, "TEXT")
		}

		// check that it's not logging JSON
		err := json.Unmarshal(logEntry, &map[string]any{})
		if !errors.As(err, new(*json.SyntaxError)) {
			t.Errorf("expected for error (%v) to be a %T", err, json.SyntaxError{})
		}
	})

	t.Run("handler nil", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, nil)

		// When called with nil handler, then it writes to the
		// package-configured logger.
		logg.SetDefaults(nil, nil)

		slog.Info(t.Name())

		if len(sink.Raw()) < 1 {
			t.Error("did not write data")
		}
	})

	t.Run("settings.ApplicationMetadata", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, &logg.Settings{
			ApplicationMetadata: []slog.Attr{slog.String("foo", "bar")},
		})

		slog.Info("settings.ApplicationMetadata")

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}

		testApplicationMetadata(t, parsedRoot, "application_metadata", map[string]string{"foo": "bar"})
	})

	t.Run("settings.ApplicationMetadataKey", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, &logg.Settings{
			ApplicationMetadata:    []slog.Attr{slog.String("branch_name", "dev"), slog.String("build_time", "now")},
			ApplicationMetadataKey: "metadata",
		})

		slog.Info("settings.ApplicationMetadataKey")

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}

		testApplicationMetadata(t, parsedRoot, "metadata", map[string]string{"branch_name": "dev", "build_time": "now"})
	})

	t.Run("settings.TraceIDKey", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, &logg.Settings{
			TraceIDKey: "id",
		})

		logg.New("trace_id").Info("settings.TraceIDKey")

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}

		testTraceID(t, parsedRoot, true, "id", "trace_id")
	})

	t.Run("settings.DataKey", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, &logg.Settings{
			DataKey: "message_data",
		})

		logg.New("", slog.String("sierra", "nevada")).Info("settings.DataKey", slog.String("foo", "bar"))

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}

		testData(t, parsedRoot, "message_data", []slog.Attr{slog.String("sierra", "nevada"), slog.String("foo", "bar")})
	})
}

func TestNew(t *testing.T) {
	setupPackageVars()
	t.Cleanup(func() { setupPackageVars() })

	t.Run("trace ID", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, nil)

		slogger := logg.New("tracing_id")

		slogger.Info("with trace ID")

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}
		testTraceID(t, parsedRoot, true, "trace_id", "tracing_id")
	})

	t.Run("without trace ID", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, nil)

		slogger := logg.New("")

		slogger.Info("without trace ID")

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}
		testTraceID(t, parsedRoot, false, "", "")
	})

	t.Run("with data attrs", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, nil)

		slogger := logg.New("", slog.String("sierra", "nevada"))

		slogger.Info("with data attrs")

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}
		testData(t, parsedRoot, "data", []slog.Attr{slog.String("sierra", "nevada")})
	})

	t.Run("initialized with data attrs and log with data attrs", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
		logg.SetDefaults(handler, nil)

		slogger := logg.New("", slog.String("sierra", "nevada"))

		slogger.Info("hello", slog.Bool("bravo", true))

		var parsedRoot map[string]any
		if err := json.Unmarshal(sink.Raw(), &parsedRoot); err != nil {
			t.Fatal(err)
		}
		testData(t, parsedRoot, "data", []slog.Attr{slog.String("sierra", "nevada"), slog.Bool("bravo", true)})
	})

	t.Run("application_metadata key not duplicated", func(t *testing.T) {
		// For this test, do not use the JSON handler and do not parse the JSON
		// because the parsing step would deduplicate the keys. This would
		// defeat the purpose of detecting duplicate keys in the raw log
		// message. Instead, inspect the data without parsing it. This test
		// assumes that the targeted values are keys. For this reason, build the
		// message as simple as necessary.
		sink := newDataSink()
		handler := slog.NewTextHandler(sink, nil)
		logg.SetDefaults(handler, &logg.Settings{
			ApplicationMetadata:    []slog.Attr{slog.String("foo", "bar")},
			ApplicationMetadataKey: "application_metadata",
		})

		alogger := logg.New("")
		alogger.Info("hello")

		rawLogEntry := string(sink.Raw())
		count := strings.Count(rawLogEntry, `application_metadata.foo=bar`)
		if count != 1 {
			t.Errorf("wrong count of substring found; got %d, expected %d", count, 1)
			t.Logf("%s", rawLogEntry)
		}

		blogger := logg.New("")
		blogger.Info("hello")

		rawLogEntry = string(sink.Raw())
		count = strings.Count(rawLogEntry, `application_metadata.foo=bar`)
		if count != 1 {
			t.Errorf("wrong count of substring found; got %d, expected %d", count, 1)
			t.Logf("%s", rawLogEntry)
		}
	})
}

func testApplicationMetadata(t *testing.T, parsedRoot map[string]any, expMetadataKey string, expData map[string]string) {
	// Check that the effects of package configuration are seen in
	// subsequent log entries.
	var parsedData map[string]any

	val, ok := parsedRoot[expMetadataKey]
	if !ok {
		t.Fatalf("expected to have key %q", expMetadataKey)
	} else if parsedData, ok = val.(map[string]any); !ok {
		t.Errorf("expected %q to be a %T", expMetadataKey, make(map[string]any))
	}

	if len(parsedData) != len(expData) {
		t.Errorf("wrong number of keys; got %d, expected %d", len(parsedData), len(expData))
	}

	for expKey, expVal := range expData {
		got, ok := parsedData[expKey]
		if !ok {
			t.Errorf("expected to have subkey [%q][%q]", expMetadataKey, expKey)
		} else if got != expVal {
			t.Errorf(
				"wrong value at [%q][%q]; got %v (type %T), expected %v (type %T)",
				expMetadataKey, expKey, got, got, expVal, expVal,
			)
		}
	}
}

func testTraceID(t *testing.T, parsedRoot map[string]any, expTraceID bool, expTraceIDKey, expTraceIDVal string) {
	if expTraceID {
		val, ok := parsedRoot[expTraceIDKey]
		if !ok {
			t.Fatalf("expected to have key %q", expTraceIDKey)
		}
		if val.(string) != expTraceIDVal {
			t.Errorf("wrong id value at %q; got %q, expected %q", expTraceIDKey, val.(string), expTraceIDVal)
		}
	} else {
		val, ok := parsedRoot[expTraceIDKey]
		if ok {
			t.Errorf("unexpected trace ID at %q; got %v", expTraceIDKey, val)
		}
	}
}

func testData(t *testing.T, parsedRoot map[string]any, expDataKey string, expData []slog.Attr) {
	if expData != nil {
		var parsedData map[string]any

		if val, ok := parsedRoot[expDataKey]; !ok {
			t.Fatalf("expected to have key %q", expDataKey)
		} else if parsedData, ok = val.(map[string]any); !ok {
			t.Fatalf("expected %q to be a %T", expDataKey, make(map[string]any))
		}

		parsedGroupAttrs := parseGroupAttrs(t, parsedData)
		testAttrs(t, parsedGroupAttrs, expData)
	} else {
		val, ok := parsedRoot[expDataKey]
		if ok {
			t.Errorf("unexpected data at %q; got %v", expDataKey, val)
		}
	}
}

// parseGroupAttrs takes the map representing the log group and approximates its
// contents into a []slog.Attr.
func parseGroupAttrs(t *testing.T, parsedGroupJSON map[string]any) []slog.Attr {
	t.Helper()

	out := make([]slog.Attr, 0, len(parsedGroupJSON))

	for key, val := range parsedGroupJSON {
		var value slog.Value
		switch v := val.(type) {
		case string:
			value = slog.StringValue(v)
		case float64: // JSON numbers are always unmarshaled as float64
			// Test logs of kind slog.KindDuration would look like a float64,
			// and would be the number of nanoseconds. Those cases are not
			// accounted for here.
			if v == float64(int(v)) {
				value = slog.IntValue(int(v))
			} else {
				value = slog.Float64Value(v)
			}
		case bool:
			value = slog.BoolValue(v)
		case time.Time:
			value = slog.TimeValue(v)
		// add more cases as needed (ie: []any for slices)
		default:
			value = slog.AnyValue(v)
		}

		out = append(out, slog.Attr{Key: key, Value: value})
	}

	return out
}

func testAttrs(t *testing.T, actual, expected []slog.Attr) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("wrong number of items; got %d, expected %d", len(actual), len(expected))
	}

	// Collect data to compare.
	actualKeyVals, expectedKeyVals := make(map[string]slog.Attr, len(actual)), make(map[string]slog.Attr, len(expected))
	var found bool
	for i := range actual {
		attr := actual[i]
		if _, found = actualKeyVals[attr.Key]; found {
			t.Fatalf("unexpected duplicate key in actual result %q", attr.Key)
		}
		actualKeyVals[attr.Key] = attr
	}
	for i := range expected {
		attr := expected[i]
		if _, found = expectedKeyVals[attr.Key]; found {
			t.Fatalf("unexpected duplicate key in expected result %q", attr.Key)
		}
		expectedKeyVals[attr.Key] = attr
	}

	// Compare keys and values.
	for i := range actual {
		actualAttr := actual[i]
		key := actualAttr.Key
		expectedAttr, found := expectedKeyVals[key]
		if !found {
			t.Errorf("unexpected key %q in actual result", key)
			continue
		}
		if !actualAttr.Equal(expectedAttr) {
			t.Errorf(
				"actual value at key %q does not equal expected value; got %v, expected %v",
				key, actualAttr.Value, expectedAttr.Value,
			)
		}
	}

	for i := range expected {
		expectedAttr := expected[i]
		key := expectedAttr.Key
		actualAttr, found := actualKeyVals[key]
		if !found {
			t.Errorf("did not find expected key %q in actual result", key)
			continue
		}
		if !actualAttr.Equal(expectedAttr) {
			t.Errorf(
				"actual value at key %q does not equal expected value; got %v, expected %v",
				key, actualAttr.Value, expectedAttr.Value,
			)
		}
	}
}

// newDataSinkAndJSONHandler creates a data sink to capture test output and a
// handler bound to that data sink.
func newDataSinkAndJSONHandler(lvl slog.Level) (sink *DataSink, handler slog.Handler) {
	sink = newDataSink()
	handler = slog.NewJSONHandler(sink, &slog.HandlerOptions{Level: lvl})
	return
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
