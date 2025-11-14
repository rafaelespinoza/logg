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
	st "github.com/rafaelespinoza/logg/slogtesting"
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
		checkGroupAttr := st.TestInGroup("application_metadata", st.TestHasAttr(slog.String("foo", "bar")))
		checkGroupAttr(t, gotAttrs)
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
		checkGroupAttrs := st.TestInGroup("metadata",
			st.TestHasAttr(slog.String("branch_name", "dev")),
			st.TestHasAttr(slog.String("build_time", "now")),
		)
		checkGroupAttrs(t, gotAttrs)
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
		checkAttr := st.TestHasAttr(slog.String("id", "trace_id"))
		checkAttr(t, gotAttrs)
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
		checkGroupAttrs := st.TestInGroup("message_data",
			st.TestHasAttr(slog.String("sierra", "nevada")),
			st.TestHasAttr(slog.String("foo", "bar")),
		)
		checkGroupAttrs(t, gotAttrs)
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
		checkAttr := st.TestHasAttr(slog.String("trace_id", "tracing_id"))
		checkAttr(t, gotAttrs)
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
		checkGroupAttr := st.TestInGroup("data", st.TestHasAttr(slog.String("sierra", "nevada")))
		checkGroupAttr(t, gotAttrs)
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
		checkGroupAttrs := st.TestInGroup("data",
			st.TestHasAttr(slog.String("sierra", "nevada")),
			st.TestHasAttr(slog.Bool("bravo", true)),
		)
		checkGroupAttrs(t, gotAttrs)
	})

	t.Run("application_metadata key not duplicated", func(t *testing.T) {
		t.Cleanup(func() { setupPackageVars() })

		run := func(t *testing.T, h slog.Handler) {
			logg.SetDefaults(h, &logg.Settings{
				ApplicationMetadata:    []slog.Attr{slog.String("foo", "bar")},
				ApplicationMetadataKey: "metadata",
			})

			alogger := logg.New("")
			alogger.Info("hello")

			blogger := logg.New("")
			blogger.Info("hello")
		}
		gotRecords := collectRecords(t, slog.LevelInfo, run)

		if len(gotRecords) != 2 {
			t.Fatalf("wrong number of record attrs; got %d, expected %d", len(gotRecords), 1)
		}

		for _, gotRecord := range gotRecords {
			checkGroupAttr := st.TestInGroup("metadata", st.TestHasAttr(slog.String("foo", "bar")))
			checkGroupAttr(t, internal.GetRecordAttrs(gotRecord))
		}
	})
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
	opts := st.AttrHandlerOptions{
		HandlerOptions: slog.HandlerOptions{Level: lvl},
		CaptureRecord:  capture,
	}
	handler := st.NewAttrHandler(&opts)
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
