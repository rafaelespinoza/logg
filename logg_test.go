package logg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rafaelespinoza/logg"
)

var pkgSink io.Writer = os.Stderr

func init() {
	setupPackageVars(pkgSink)
}

// setupPackageVars sets up some package-level state expected by most tests.
// Output for package-level logging functions will be written to w.
func setupPackageVars(w io.Writer) {
	// versioningData should be present in all subsequent log entries; not only
	// from package-level functions, but also from Emitter events.
	versioningData := []slog.Attr{slog.String("branch_name", "dev"), slog.String("build_time", "now")}

	// Tests rely on parsing the JSON log entries to check for correctness.
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
	logg.Setup(handler, versioningData...)
}

func TestSetup(t *testing.T) {
	t.Cleanup(func() { setupPackageVars(pkgSink) }) // Restore package-level variables for remainder of test suite.

	t.Run("level", func(t *testing.T) {
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
				sink, handler := newDataSinkAndJSONHandler(test.inputLevel)
				logg.Setup(handler)

				logg.Debug(t.Name())
				gotDebug := sink.Raw()
				if test.expDebugOutput && len(gotDebug) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelDebug.String())
				} else if !test.expInfoOutput && len(gotDebug) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelDebug.String())
				}

				logg.Info(t.Name())
				gotInfo := sink.Raw()
				if test.expInfoOutput && len(gotInfo) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelInfo.String())
				} else if !test.expInfoOutput && len(gotInfo) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelInfo.String())
				}

				logg.Warn(t.Name())
				gotWarn := sink.Raw()
				if test.expWarnOutput && len(gotWarn) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelWarn.String())
				} else if !test.expWarnOutput && len(gotWarn) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelWarn.String())
				}

				logg.Error(errors.New("test"), t.Name())
				gotError := sink.Raw()
				if test.expErrorOutput && len(gotError) < 1 {
					t.Errorf("expected to write data at level %s", slog.LevelError.String())
				} else if !test.expErrorOutput && len(gotError) > 0 {
					t.Errorf("unexpected data written at level %s", slog.LevelError.String())
				}
			})
		}
	})

	t.Run("format TEXT", func(t *testing.T) {
		sink := newDataSink()
		handler := slog.NewTextHandler(sink, &slog.HandlerOptions{Level: slog.LevelInfo})
		logg.Setup(handler)

		logg.Info(t.Name())

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
}

func TestDebug(t *testing.T) {
	t.Cleanup(func() { setupPackageVars(pkgSink) }) // Restore package-level variables for remainder of test suite.

	t.Run("no attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		logg.Debug("hello debug")

		testLogg(t, sink.Raw(), nil, "hello debug", false, "", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		logg.Debug("hello debug", slog.String("foo", "bar"))

		testLogg(t, sink.Raw(), nil, "hello debug", false, "", []slog.Attr{
			slog.String("foo", "bar"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestInfo(t *testing.T) {
	t.Cleanup(func() { setupPackageVars(pkgSink) }) // Restore package-level variables for remainder of test suite.

	t.Run("no attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		logg.Info("hello info")

		testLogg(t, sink.Raw(), nil, "hello info", false, "", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		logg.Info("hello info", slog.String("foo", "bar"))

		testLogg(t, sink.Raw(), nil, "hello info", false, "", []slog.Attr{
			slog.String("foo", "bar"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestWarn(t *testing.T) {
	t.Cleanup(func() { setupPackageVars(pkgSink) }) // Restore package-level variables for remainder of test suite.

	t.Run("no attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		logg.Warn("hello warn")

		testLogg(t, sink.Raw(), nil, "hello warn", false, "", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		logg.Warn("hello warn", slog.String("foo", "bar"))

		testLogg(t, sink.Raw(), nil, "hello warn", false, "", []slog.Attr{
			slog.String("foo", "bar"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestError(t *testing.T) {
	t.Cleanup(func() { setupPackageVars(pkgSink) }) // Restore package-level variables for remainder of test suite.

	t.Run("no attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		err := errors.New("OOF")
		logg.Error(err, "hello error")

		testLogg(t, sink.Raw(), err, "hello error", false, "", nil)
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		sink := newDataSink()
		setupPackageVars(sink)
		err := errors.New("OOF")
		logg.Error(err, "hello error", slog.String("bar", "foo"))

		testLogg(t, sink.Raw(), err, "hello error", false, "", []slog.Attr{
			slog.String("bar", "foo"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestLogg(t *testing.T) {
	setupPackageVars(pkgSink)

	t.Run("Debug", func(t *testing.T) {
		t.Run("from logger", func(t *testing.T) {
			// setup
			sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			// action
			logger.Debug("hello")

			// test
			testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("from logger.WithData", func(t *testing.T) {
			// setup
			sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			// action
			logger.WithData([]slog.Attr{
				slog.Bool("bravo", true),
				slog.String("delta", (234 * time.Millisecond).String()),
				slog.Float64("foxtrot", 1.23),
				slog.Int("india", 10),
			}).Debug("goodbye")

			// test
			testLogg(t, sink.Raw(), nil, "goodbye", false, "", []slog.Attr{
				slog.Bool("bravo", true),
				slog.String("delta", (234 * time.Millisecond).String()),
				slog.Float64("foxtrot", 1.23),
				slog.Int("india", 10),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("another event from same logger can have its own fields", func(t *testing.T) {
			// setup
			sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
			logger := logg.New(handler, slog.String("sierra", "nevada"))
			logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Debug("goodbye")

			// action
			logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Debug("goodbye again")

			// test
			testLogg(t, sink.Raw(), nil, "goodbye again", false, "", []slog.Attr{
				slog.Bool("zulu", true),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("with attrs", func(t *testing.T) {
			// setup
			sink, handler := newDataSinkAndJSONHandler(slog.LevelDebug)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			// action
			logger.Debug("hello", slog.Int("i", 1))

			// test
			testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{
				slog.String("sierra", "nevada"),
				slog.Int("i", 1),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}

			// Calling a log method with attrs does not affect logger's fields
			logger.Debug("hello again")

			testLogg(t, sink.Raw(), nil, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("not enabled", func(t *testing.T) {
			// setup
			sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
			logger := logg.New(handler)

			// action
			logger.Debug("hello")

			// test
			if len(sink.Raw()) > 0 {
				t.Errorf("unexpected data written at level %s", slog.LevelDebug.String())
			}
		})
	})

	t.Run("Info", func(t *testing.T) {
		t.Run("from logger", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.Info("hello")

			testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("from logger.WithData", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.WithData([]slog.Attr{
				slog.Bool("bravo", true),
				slog.String("delta", (234 * time.Millisecond).String()),
				slog.Float64("foxtrot", 1.23),
				slog.Int("india", 10),
			}).Info("goodbye")

			testLogg(t, sink.Raw(), nil, "goodbye", false, "", []slog.Attr{
				// The types of expected values are changed due to the way
				// encoding/json unmarshals any values. This is far from
				// perfect, so at this point, considering the field tests to be
				// "close enough".
				slog.Bool("bravo", true),
				slog.String("delta", (234 * time.Millisecond).String()),
				slog.Float64("foxtrot", 1.23),
				slog.Int("india", 10),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("another event from same logger can have its own fields", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))
			logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Info("goodbye")

			logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Info("goodbye again")

			testLogg(t, sink.Raw(), nil, "goodbye again", false, "", []slog.Attr{
				slog.Bool("zulu", true),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("with attrs", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.Info("hello", slog.Int("i", 1))

			testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{
				slog.String("sierra", "nevada"),
				slog.Int("i", 1),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}

			// Calling a log method with attrs does not affect logger's fields
			logger.Info("hello again")

			testLogg(t, sink.Raw(), nil, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("not enabled", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
			logger := logg.New(handler)

			logger.Info("hello")

			if len(sink.Raw()) > 0 {
				t.Errorf("unexpected data written at level %s", slog.LevelInfo.String())
			}
		})
	})

	t.Run("Warn", func(t *testing.T) {
		t.Run("from logger", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.Warn("hello")

			testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("from logger.WithData", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.WithData([]slog.Attr{
				slog.Bool("bravo", true),
				slog.String("delta", (234 * time.Millisecond).String()),
				slog.Float64("foxtrot", 1.23),
				slog.Int("india", 10),
			}).Warn("goodbye")

			testLogg(t, sink.Raw(), nil, "goodbye", false, "", []slog.Attr{
				slog.Bool("bravo", true),
				slog.String("delta", (234 * time.Millisecond).String()),
				slog.Float64("foxtrot", 1.23),
				slog.Int("india", 10),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("another event from same logger can have its own fields", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))
			logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Warn("goodbye")

			logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Warn("goodbye again")

			testLogg(t, sink.Raw(), nil, "goodbye again", false, "", []slog.Attr{
				slog.Bool("zulu", true),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("with attrs", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.Warn("hello", slog.Int("i", 1))

			testLogg(t, sink.Raw(), nil, "hello", false, "", []slog.Attr{
				slog.String("sierra", "nevada"),
				slog.Int("i", 1),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}

			// Calling a log method with attrs does not affect logger's fields
			logger.Warn("hello again")

			testLogg(t, sink.Raw(), nil, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("not enabled", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
			logger := logg.New(handler)

			logger.Warn("hello")

			if len(sink.Raw()) > 0 {
				t.Errorf("unexpected data written at level %s", slog.LevelWarn.String())
			}
		})
	})

	t.Run("Error", func(t *testing.T) {
		t.Run("from logger", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.Error(errors.New("hello"), "logger error")

			testLogg(t, sink.Raw(), errors.New("hello"), "logger error", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("from logger.WithData", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.WithData([]slog.Attr{
				slog.Bool("bravo", false),
				slog.String("delta", (432 * time.Millisecond).String()),
				slog.Float64("foxtrot", 3.21),
				slog.Int("india", 100),
			}).Error(errors.New("goodbye"), "event error")

			testLogg(t, sink.Raw(), errors.New("goodbye"), "event error", false, "", []slog.Attr{
				// The types of expected values are changed due to the way
				// encoding/json unmarshals any values. This is far from
				// perfect, so at this point, considering the field tests to be
				// "close enough".
				slog.Bool("bravo", false),
				slog.String("delta", (432 * time.Millisecond).String()),
				slog.Float64("foxtrot", 3.21),
				slog.Int("india", 100),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("another event from same logger can have its own fields", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))
			logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Info("goodbye")

			logger.WithData([]slog.Attr{slog.Bool("zulu", true)}).Error(errors.New("bye"), "goodbye again")

			testLogg(t, sink.Raw(), errors.New("bye"), "goodbye again", false, "", []slog.Attr{
				slog.Bool("zulu", true),
				slog.String("sierra", "nevada"),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("with attrs", func(t *testing.T) {
			err := errors.New("test")
			sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)
			logger := logg.New(handler, slog.String("sierra", "nevada"))

			logger.Error(err, "hello", slog.Int("i", 1))

			testLogg(t, sink.Raw(), err, "hello", false, "", []slog.Attr{
				slog.String("sierra", "nevada"),
				slog.Int("i", 1),
			})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}

			// Calling a log method with attrs does not affect logger's fields
			logger.Error(err, "hello again")

			testLogg(t, sink.Raw(), err, "hello again", false, "", []slog.Attr{slog.String("sierra", "nevada")})
			if t.Failed() {
				t.Logf("%s", sink.Raw())
			}
		})

		t.Run("not enabled", func(t *testing.T) {
			sink, handler := newDataSinkAndJSONHandler(slog.LevelError + 1)
			logger := logg.New(handler)

			logger.Error(errors.New("test"), "hello")

			if len(sink.Raw()) > 0 {
				t.Errorf("unexpected data written at level %s", slog.LevelError.String())
			}
		})
	})
}

func TestWithID(t *testing.T) {
	t.Run("logger passes ID to event", func(t *testing.T) {
		// The logger calls WithID, but the event does not.
		ctx := logg.SetID(context.Background(), "logger_id")
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler, slog.String("sierra", "nevada")).WithID(ctx)
		logger.Info("logger with id")
		testLogg(t, sink.Raw(), nil, "logger with id", true, "logger_id", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).Info("event with id")
		testLogg(t, sink.Raw(), nil, "event with id", true, "logger_id", []slog.Attr{
			slog.String("sierra", "nevada"),
			slog.Bool("bravo", true),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("event can set its own ID", func(t *testing.T) {
		// The logger does not call WithID, but the event does.
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler, slog.String("sierra", "nevada"))
		logger.Info("logger without id")
		testLogg(t, sink.Raw(), nil, "logger without id", false, "", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		ctx := logg.SetID(context.Background(), "event_id")
		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).WithID(ctx).Info("event with own id")
		testLogg(t, sink.Raw(), nil, "event with own id", true, "event_id", []slog.Attr{
			slog.Bool("bravo", true),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("event can set its own ID despite logger setting an ID", func(t *testing.T) {
		// The logger calls WithID and so does the event.

		ctxA := logg.SetID(context.Background(), "logger_id")
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler, slog.String("sierra", "nevada")).WithID(ctxA)
		logger.Info("logger with id")
		testLogg(t, sink.Raw(), nil, "logger with id", true, "logger_id", []slog.Attr{slog.String("sierra", "nevada")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// derive context from previous context, but set another ID.
		ctxB := logg.SetID(ctxA, "event_id")
		logger.WithData([]slog.Attr{slog.Bool("bravo", true)}).WithID(ctxB).Info("event with own id")
		testLogg(t, sink.Raw(), nil, "event with own id", true, "event_id", []slog.Attr{
			slog.Bool("bravo", true),
			slog.String("sierra", "nevada"),
		})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func TestWithData(t *testing.T) {
	t.Run("allows input fields to replace existing fields", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler, slog.String("foo", "alfa"), slog.Bool("bar", true))
		logger.Info("a")
		testLogg(t, sink.Raw(), nil, "a", false, "", []slog.Attr{slog.String("foo", "alfa"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		event := logger.WithData([]slog.Attr{slog.String("foo", "bravo")})
		event.Info("b")
		testLogg(t, sink.Raw(), nil, "b", false, "", []slog.Attr{slog.String("foo", "bravo"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		otherEvent := event.WithData([]slog.Attr{slog.Bool("bar", false)})
		otherEvent.Info("c")
		testLogg(t, sink.Raw(), nil, "c", false, "", []slog.Attr{slog.String("foo", "bravo"), slog.Bool("bar", false)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// check that first logger's original data hasn't changed unexpectedly.
		logger.Info("d")
		testLogg(t, sink.Raw(), nil, "d", false, "", []slog.Attr{slog.String("foo", "alfa"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}

		// check that first event's original data hasn't changed unexpectedly.
		event.Info("e")
		testLogg(t, sink.Raw(), nil, "e", false, "", []slog.Attr{slog.String("foo", "bravo"), slog.Bool("bar", true)})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("when initial state of fields is nil", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler).WithData([]slog.Attr{slog.String("foo", "alfa")})
		logger.Info("a")
		testLogg(t, sink.Raw(), nil, "a", false, "", []slog.Attr{slog.String("foo", "alfa")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})

	t.Run("when passed nil", func(t *testing.T) {
		sink, handler := newDataSinkAndJSONHandler(slog.LevelInfo)

		logger := logg.New(handler, slog.String("foo", "alfa")).WithData(nil)
		logger.Info("a")
		testLogg(t, sink.Raw(), nil, "a", false, "", []slog.Attr{slog.String("foo", "alfa")})
		if t.Failed() {
			t.Logf("%s", sink.Raw())
		}
	})
}

func testLogg(t *testing.T, in []byte, expErr error, expMessage string, expTraceIDKey bool, expTraceIDVal string, expData []slog.Attr) {
	t.Helper()

	const (
		errorKey   = "error"
		messageKey = "msg"
		versionKey = "version"
		traceIDKey = "x_trace_id"
		dataKey    = "data"
	)

	var parsedRoot map[string]any

	if err := json.Unmarshal(in, &parsedRoot); err != nil {
		t.Fatal(err)
	}

	t.Run(errorKey+" field", func(t *testing.T) {
		if expErr != nil {
			val, ok := parsedRoot[errorKey]
			if !ok {
				t.Fatalf("expected to have key %q", errorKey)
			}
			if val.(string) != expErr.Error() {
				t.Errorf("wrong value at %q; got %q, expected %q", errorKey, val.(string), expErr.Error())
			}
		} else {
			val, ok := parsedRoot[errorKey]
			if ok {
				t.Errorf("unexpected error at %q; got %v", errorKey, val)
			}
		}
	})

	t.Run(messageKey+" field", func(t *testing.T) {
		val, ok := parsedRoot[messageKey]
		if !ok {
			t.Fatalf("expected to have key %q", messageKey)
		}
		if val.(string) != expMessage {
			t.Errorf("wrong value at %q; got %q, expected %q", messageKey, val.(string), expMessage)
		}
	})

	t.Run(versionKey+" field", func(t *testing.T) {
		// Check that the effects of package configuration are seen in
		// subsequent log entries.
		var parsedVersioningData map[string]any

		val, ok := parsedRoot[versionKey]
		if !ok {
			t.Fatalf("expected to have key %q", versionKey)
		} else if parsedVersioningData, ok = val.(map[string]any); !ok {
			t.Errorf("expected %q to be a %T", versionKey, make(map[string]any))
		}

		expVersioningData := map[string]string{"branch_name": "dev", "build_time": "now"}
		if len(parsedVersioningData) != len(expVersioningData) {
			t.Errorf("wrong number of keys; got %d, expected %d", len(parsedVersioningData), len(expVersioningData))
		}

		for expKey, expVal := range expVersioningData {
			got, ok := parsedVersioningData[expKey]
			if !ok {
				t.Errorf("expected to have subkey [%q][%q]", versionKey, expKey)
			} else if got != expVal {
				t.Errorf(
					"wrong value at [%q][%q]; got %v (type %T), expected %v (type %T)",
					versionKey, expKey, got, got, expVal, expVal,
				)
			}
		}
	})

	t.Run(traceIDKey+" field", func(t *testing.T) {
		numTraceKeyValues := strings.Count(string(in), traceIDKey)
		if expTraceIDKey {
			val, ok := parsedRoot[traceIDKey]
			if !ok {
				t.Fatalf("expected to have key %q", traceIDKey)
			}
			if val.(string) != expTraceIDVal {
				t.Errorf("wrong id value at %q; got %q, expected %q", traceIDKey, val.(string), expTraceIDVal)
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
	})

	t.Run(dataKey+" field", func(t *testing.T) {
		if expData != nil {
			var parsedData map[string]any

			if val, ok := parsedRoot[dataKey]; !ok {
				t.Fatalf("expected to have key %q", dataKey)
			} else if parsedData, ok = val.(map[string]any); !ok {
				t.Fatalf("expected %q to be a %T", dataKey, make(map[string]any))
			}

			parsedGroupAttrs := parseGroupAttrs(t, parsedData)
			testAttrs(t, parsedGroupAttrs, expData)
		} else {
			val, ok := parsedRoot[dataKey]
			if ok {
				t.Errorf("unexpected data at %q; got %v", dataKey, val)
			}
		}
	})
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
