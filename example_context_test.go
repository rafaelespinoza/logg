package logg_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/rafaelespinoza/logg"
)

// Use values of these types as keys for accessing values on a context. It's
// best practice to use unexported types to reduce the likelihood of collisions.
type (
	traceContextKey  struct{}
	loggerContextKey struct{}
)

// SetContextLoggerWithID creates a new logger with a tracing ID and dataAttrs,
// and returns a new context with the logger and tracing ID. Retrieve the logger
// with GetContextLogger.
func SetContextLoggerWithID(ctx context.Context, traceID string, dataAttrs ...slog.Attr) context.Context {
	lgr := logg.New(traceID, dataAttrs...)
	outCtx := context.WithValue(ctx, loggerContextKey{}, lgr)
	outCtx = context.WithValue(outCtx, traceContextKey{}, traceID)
	return outCtx
}

// GetContextLogger fetches the logger on the context, which would have been set
// by SetContextLoggerWithID. If no logger is found, then it returns the default
// logger from the slog package.
func GetContextLogger(ctx context.Context) *slog.Logger {
	lgr, found := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if found {
		return lgr
	}
	return slog.Default()
}

// A demo of context-based loggers. For brevity's sake, this example lacks some
// things you may want in a real application, such as using the context to
// transmit an existing tracing ID or generating a tracing ID in case the
// context doesn't have one. That exercise is up to you.
func Example_context() {
	defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// Remove the time key so that output is consistent.
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) < 1 && a.Key == slog.TimeKey {
				a = slog.Attr{}
			}
			return a
		},
	})
	logg.SetDefaults(defaultHandler, &logg.Settings{
		ApplicationMetadata:    []slog.Attr{slog.String("branch", "main")},
		ApplicationMetadataKey: "version",
	})

	aCtx := SetContextLoggerWithID(context.Background(), "unique_id")
	aLogger := GetContextLogger(aCtx)
	aLogger.Info("Avalanches in Atlanta")

	bCtx := SetContextLoggerWithID(context.Background(), "example_with_data_attrs", slog.String("foo", "bar"))
	bLogger := GetContextLogger(bCtx)
	bLogger.Info("Blizzards in Bakersfield")
	// Output:
	// level=INFO msg="Avalanches in Atlanta" version.branch=main trace_id=unique_id
	// level=INFO msg="Blizzards in Bakersfield" version.branch=main trace_id=example_with_data_attrs data.foo=bar
}
