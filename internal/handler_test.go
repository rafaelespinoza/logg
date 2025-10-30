package internal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/rafaelespinoza/logg/internal"
)

func TestSlogtest(t *testing.T) {
	originalDefaults := slog.Default()
	t.Cleanup(func() { slog.SetDefault(originalDefaults) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				attr = slog.Attr{}
			}
			return attr
		},
	})))

	var buf bytes.Buffer
	captureJSONRecord := makeJSONRecordCapturer(&buf)
	newHandler := func(t *testing.T) slog.Handler {
		t.Helper()
		buf.Reset() // try to make each test a clean slate.
		opts := internal.AttrHandlerOptions{
			HandlerOptions: slog.HandlerOptions{Level: slog.LevelInfo},
			CaptureRecord:  captureJSONRecord,
		}
		return internal.NewAttrHandler(&opts)
	}
	makeTestResults := func(t *testing.T) (out map[string]any) {
		t.Helper()
		line := buf.Bytes()
		if len(line) == 0 {
			return
		}
		t.Logf("%s", line)
		if err := json.Unmarshal(line, &out); err != nil {
			t.Fatal(err)
		}
		return
	}

	slogtest.Run(t, newHandler, makeTestResults)
}

func TestAttrHandler(t *testing.T) {
	tests := []struct {
		name   string
		opts   *internal.AttrHandlerOptions
		action func(*testing.T, slog.Handler)
		expect func(*testing.T, []slog.Record)
	}{
		{
			name: "options.ReplaceAttr time key",
			opts: &internal.AttrHandlerOptions{
				HandlerOptions: slog.HandlerOptions{
					ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
						if a.Key == slog.TimeKey {
							a.Key = "timestamp"
						}
						return a
					},
				},
			},
			action: func(t *testing.T, h slog.Handler) {
				rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
				err := h.Handle(context.Background(), rec)
				if err != nil {
					t.Error(err)
				}
			},
			expect: func(t *testing.T, got []slog.Record) {
				checkResultsLength(t, got, 1)
				if t.Failed() {
					return
				}
				attrs := internal.GetRecordAttrs(got[0])
				checkForAttrWithKey(t, attrs, "timestamp")
			},
		},
		{
			name: "options.Level level key",
			opts: &internal.AttrHandlerOptions{
				HandlerOptions: slog.HandlerOptions{
					ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
						if a.Key == slog.LevelKey {
							a.Key = "sev"
						}
						return a
					},
				},
			},
			action: func(t *testing.T, h slog.Handler) {
				rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
				err := h.Handle(context.Background(), rec)
				if err != nil {
					t.Error(err)
				}
			},
			expect: func(t *testing.T, got []slog.Record) {
				checkResultsLength(t, got, 1)
				if t.Failed() {
					return
				}
				attrs := internal.GetRecordAttrs(got[0])
				checkForAttrWithKey(t, attrs, "sev")
			},
		},
		{
			name: "options.AddSource",
			opts: &internal.AttrHandlerOptions{
				HandlerOptions: slog.HandlerOptions{AddSource: true},
			},
			action: func(t *testing.T, h slog.Handler) {
				rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 1)
				h.Handle(context.Background(), rec)
			},
			expect: func(t *testing.T, got []slog.Record) {
				checkResultsLength(t, got, 1)
				if t.Failed() {
					return
				}
				const targetKey = "source"
				targetAttrs := make([]slog.Attr, 0, 1)
				for _, attr := range internal.GetRecordAttrs(got[0]) {
					if attr.Key == targetKey {
						targetAttrs = append(targetAttrs, attr)
					}
				}
				if len(targetAttrs) != 1 {
					t.Fatalf("wrong number of attributes with key %q; got %d, expected %d", targetKey, len(targetAttrs), 1)
				}
				if targetAttrs[0].Value.Any() == nil {
					t.Errorf("%s attribute value should be non-nil", targetKey)
				}
			},
		},
		{
			name: "initialized with empty options",
			opts: nil,
			action: func(t *testing.T, h slog.Handler) {
				rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
				err := h.Handle(context.Background(), rec)
				if err != nil {
					t.Error(err)
				}
			},
			expect: func(t *testing.T, got []slog.Record) {
				checkResultsLength(t, got, 0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got []slog.Record
			if test.opts != nil {
				test.opts.CaptureRecord = func(r slog.Record) error {
					got = append(got, r)
					return nil
				}
			}

			h := internal.NewAttrHandler(test.opts)
			test.action(t, h)

			test.expect(t, got)
		})
	}
}

func TestAttrHandlerEnabled(t *testing.T) {
	tests := []struct {
		name              string
		handlerOptionsNil bool
		handlerLevel      slog.Level
		expDebugOutput    bool
		expInfoOutput     bool
		expWarnOutput     bool
		expErrorOutput    bool
	}{
		{
			handlerLevel:   slog.LevelDebug,
			expDebugOutput: true, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
		},
		{
			handlerLevel:   slog.LevelInfo,
			expDebugOutput: false, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
		},
		{
			handlerLevel:   slog.LevelWarn,
			expDebugOutput: false, expInfoOutput: false, expWarnOutput: true, expErrorOutput: true,
		},
		{
			handlerLevel:   slog.LevelError,
			expDebugOutput: false, expInfoOutput: false, expWarnOutput: false, expErrorOutput: true,
		},
		{
			handlerLevel:   slog.LevelError + 1,
			expDebugOutput: false, expInfoOutput: false, expWarnOutput: false, expErrorOutput: false,
		},
		{
			name:              "handler options nil",
			handlerOptionsNil: true,
			expDebugOutput:    false, expInfoOutput: true, expWarnOutput: true, expErrorOutput: true,
		},
	}

	runTest := func(t *testing.T, optsNil bool, handlerLevel, recordLevel slog.Level, expOutput bool) {
		var opts *internal.AttrHandlerOptions
		if !optsNil {
			opts = &internal.AttrHandlerOptions{
				HandlerOptions: slog.HandlerOptions{Level: handlerLevel},
			}
		}
		handler := internal.NewAttrHandler(opts)
		enabled := handler.Enabled(context.Background(), recordLevel)
		if expOutput && !enabled {
			t.Errorf("expected data at level %s", recordLevel.String())
		} else if !expOutput && enabled {
			t.Errorf("unexpected data at level %s", recordLevel.String())
		}
	}

	for _, test := range tests {
		name := test.name
		if name == "" {
			name = "handler level " + test.handlerLevel.String()
		}

		t.Run(name, func(t *testing.T) {
			t.Run("record "+slog.LevelDebug.String(), func(t *testing.T) {
				runTest(t, test.handlerOptionsNil, test.handlerLevel, slog.LevelDebug, test.expDebugOutput)
			})

			t.Run("record "+slog.LevelInfo.String(), func(t *testing.T) {
				runTest(t, test.handlerOptionsNil, test.handlerLevel, slog.LevelInfo, test.expInfoOutput)
			})

			t.Run("record "+slog.LevelWarn.String(), func(t *testing.T) {
				runTest(t, test.handlerOptionsNil, test.handlerLevel, slog.LevelWarn, test.expWarnOutput)
			})

			t.Run("record "+slog.LevelError.String(), func(t *testing.T) {
				runTest(t, test.handlerOptionsNil, test.handlerLevel, slog.LevelError, test.expErrorOutput)
			})
		})
	}
}

func TestAttrHandlerNoCapture(t *testing.T) {
	// Sanity check that a panic does not occur when a Handler is initialized
	// without a record capture callback.
	defer func() {
		if got := recover(); got != nil {
			t.Errorf("handler.Handle should not panic if its records capture callback is nil %v", got)
		}
	}()

	h := internal.NewAttrHandler(&internal.AttrHandlerOptions{})
	err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "m", 0))
	if err != nil {
		t.Error(err)
	}
}

func checkResultsLength[T any](t *testing.T, got []T, expLen int) {
	t.Helper()
	if len(got) != expLen {
		t.Errorf("wrong number of results; got %d, expected %d", len(got), expLen)
	}
}

func checkForAttrWithKey(t *testing.T, attrs []slog.Attr, targetKey string) {
	t.Helper()
	for _, attr := range attrs {
		if attr.Key == targetKey {
			return
		}
	}
	t.Errorf("did not find expected key %s", targetKey)
}
