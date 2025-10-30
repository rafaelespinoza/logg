package logg_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/rafaelespinoza/logg"
	"github.com/rafaelespinoza/logg/internal"
)

// setupPackageVars sets up some package-level state expected by most tests.
// Output will be written to os.Stderr.
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

				sink := newDataSink()
				opts := slog.HandlerOptions{Level: test.inputLevel}
				handler := slog.NewTextHandler(sink, &opts)
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

	t.Run("handler format JSON", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		sink := newDataSink()
		handler := slog.NewJSONHandler(sink, &slog.HandlerOptions{Level: slog.LevelInfo})
		logg.SetDefaults(handler, nil)

		slog.Info(t.Name())

		logEntry := sink.Raw()
		if !strings.Contains(string(logEntry), "JSON") {
			t.Fatalf("expected output %q to contain %q", logEntry, "JSON")
		}

		// check that it's logging JSON
		err := json.Unmarshal(logEntry, &map[string]any{})
		if err != nil {
			t.Errorf("output should be json %v", err)
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

		run := func(t *testing.T, h slog.Handler) {
			logg.SetDefaults(h, nil)

			// When called with nil handler, then it writes to the
			// package-configured logger.
			logg.SetDefaults(nil, nil)
			slog.Info(t.Name())
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
	})

	t.Run("settings.ApplicationMetadata", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, handler slog.Handler) {
			logg.SetDefaults(handler, &logg.Settings{
				ApplicationMetadata: []slog.Attr{slog.String("foo", "bar")},
			})
			slog.Info("settings.ApplicationMetadata")
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testGroupAttr(t, gotAttrs, internal.SlogGroupAttrs("application_metadata", slog.String("foo", "bar")))
	})

	t.Run("settings.ApplicationMetadataKey", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, handler slog.Handler) {
			logg.SetDefaults(handler, &logg.Settings{
				ApplicationMetadata:    []slog.Attr{slog.String("branch_name", "dev"), slog.String("build_time", "now")},
				ApplicationMetadataKey: "metadata",
			})
			slog.Info("settings.ApplicationMetadataKey")
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testGroupAttr(t, gotAttrs, internal.SlogGroupAttrs("metadata",
			slog.String("branch_name", "dev"),
			slog.String("build_time", "now"),
		))
	})

	t.Run("settings.TraceIDKey", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, handler slog.Handler) {
			logg.SetDefaults(handler, &logg.Settings{TraceIDKey: "id"})
			logg.New("trace_id").Info("settings.TraceIDKey")
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testTraceIDAttr(t, gotAttrs, slog.String("id", "trace_id"))
	})

	t.Run("settings.DataKey", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, handler slog.Handler) {
			logg.SetDefaults(handler, &logg.Settings{
				DataKey: "message_data",
			})
			logg.New("", slog.String("sierra", "nevada")).Info("settings.DataKey", slog.String("foo", "bar"))
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testGroupAttr(t, gotAttrs, internal.SlogGroupAttrs("message_data",
			slog.String("sierra", "nevada"),
			slog.String("foo", "bar")),
		)
	})
}

func TestNew(t *testing.T) {
	setupPackageVars()
	t.Cleanup(func() { setupPackageVars() })

	t.Run("trace ID", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, h slog.Handler) {
			logg.SetDefaults(h, nil)
			slogger := logg.New("tracing_id")
			slogger.Info("with trace ID")
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testTraceIDAttr(t, gotAttrs, slog.String("trace_id", "tracing_id"))
	})

	t.Run("with data attrs", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, h slog.Handler) {
			logg.SetDefaults(h, nil)
			slogger := logg.New("", slog.String("sierra", "nevada"))
			slogger.Info("with data attrs")
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testGroupAttr(t, gotAttrs, internal.SlogGroupAttrs("data", slog.String("sierra", "nevada")))
	})

	t.Run("log with data attrs", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, h slog.Handler) {
			logg.SetDefaults(h, nil)
			slogger := logg.New("", slog.String("sierra", "nevada"))
			slogger.Info("hello", slog.Bool("bravo", true))
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 1 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}
		gotAttrs := internal.GetRecordAttrs(gotRecords[0])
		testGroupAttr(t, gotAttrs, internal.SlogGroupAttrs("data",
			slog.String("sierra", "nevada"),
			slog.Bool("bravo", true),
		))
	})

	t.Run("application_metadata key not duplicated", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

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

func testGroupAttr(t *testing.T, gotAttrs []slog.Attr, expectedGroup slog.Attr) {
	t.Helper()
	if got := expectedGroup.Value.Kind(); got != slog.KindGroup {
		t.Fatalf("test setup error, the expected attribute must be %q; got %q", slog.KindGroup, got)
	}

	targetAttrs := make([]slog.Attr, 0, 1)
	for _, attr := range gotAttrs {
		if attr.Key == expectedGroup.Key {
			targetAttrs = append(targetAttrs, attr)
		}
	}
	if len(targetAttrs) != 1 {
		t.Fatalf("wrong number of attributes found; got %d, expected %d", len(targetAttrs), 1)
	}

	gotAttr := targetAttrs[0]
	if got := gotAttr.Value.Kind(); got != slog.KindGroup {
		t.Fatalf("unexpected Kind for attribute; got %q, expected %q", got, slog.KindGroup)
	}

	gotGroup, expGroup := gotAttr.Value.Group(), expectedGroup.Value.Group()
	testAttrs(t, gotGroup, expGroup)
}

func testTraceIDAttr(t *testing.T, gotAttrs []slog.Attr, expected slog.Attr) {
	t.Helper()

	targetAttrs := make([]slog.Attr, 0, 1)
	for _, attr := range gotAttrs {
		if attr.Key == expected.Key {
			targetAttrs = append(targetAttrs, attr)
		}
	}
	if len(targetAttrs) != 1 {
		t.Fatalf("wrong number of trace ID attributes; got %d, expected %d", len(targetAttrs), 1)
	}

	gotTraceIDAttr := targetAttrs[0]
	got := gotTraceIDAttr.Value.String()
	exp := expected.Value.String()
	if got != exp {
		t.Errorf("wrong trace ID value at key %q; got %q, expected %q", expected.Key, got, exp)
	}
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

// collectRecords calls the run function to exercise the code to test and
// returns the []slog.Record passed to each successive invocation of the
// slog.Handler.Handle method for cases where the Handler is Enabled. To access
// the attributes of each record, call internal.GetRecordAttrs.
//
// If the test only needs to look for non-zero (ie: did it log or not?), or do
// some very simple string matching, then the newDataSink function and the
// DataSink type are probably going to be easier. Use this function when the
// test needs a more detailed look at the overall shape of the logging entry and
// the attributes contained within.
func collectRecords(t *testing.T, lvl slog.Level, run func(*testing.T, slog.Handler)) (out []slog.Record) {
	capture := func(r slog.Record) error {
		out = append(out, r)
		return nil
	}
	opts := internal.AttrHandlerOptions{
		HandlerOptions: slog.HandlerOptions{Level: lvl},
		CaptureRecord:  capture,
	}
	handler := internal.NewAttrHandler(&opts)
	run(t, handler)
	return out
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
