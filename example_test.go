package logg_test

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rafaelespinoza/logg"
)

func ExampleConfigure() {
	logg.Configure(
		os.Stderr,
		map[string]string{"branch_name": "main", "build_time": "20060102T150415", "commit_hash": "deadbeef"},
	)
}

func ExampleConfigure_multipleSinks() {
	// Write to standard error as well as some file.
	var file io.Writer

	logg.Configure(
		os.Stderr,
		map[string]string{"branch_name": "main", "build_time": "20060102T150415", "commit_hash": "1337d00d"},
		file,
	)
}

func ExampleConfigure_possiblyUnintendedConfiguration() {
	// This example shows how the first usage of logg, that is not Configure,
	// may unintentionally set up your logger.
	logg.Infof("these writes go")
	logg.Infof("to standard error")
	logg.Infof("by default")

	// Then your code attempts to configure the logger to write to someSocket.
	var someSocket io.Writer
	logg.Configure(
		someSocket,
		map[string]string{"branch_name": "main", "build_time": "20060102T150415", "commit_hash": "feedface"},
	)

	// Whoops, these logging event will continue to go to standard error. This
	// may not be what you want. The solution would be to call Configure before
	// emitting any kind of event.
	logg.Infof("hello, is there")
	logg.Infof("anybody out there?")
}

func ExampleCtxWithID() {
	// Create a new context to ensure the same tracing ID on each subsequent
	// tracing event.
	ctxA := logg.CtxWithID(context.Background())
	alfa := logg.New(map[string]interface{}{"a": "A"}).WithID(ctxA)
	alfa.Infof("altoona")
	alfa.Infof("athletic")

	// Attempting to create a logger using the same context would yield the same
	// tracing ID on each event.
	ctxB := logg.CtxWithID(ctxA)
	bravo := logg.New(map[string]interface{}{"b": "B"}).WithID(ctxB)
	bravo.Infof("boston")
	bravo.Infof("boisterous")

	// If you need another tracing ID, then use a brand-new context as the
	// parent context create create another context (using a
	// brand-new context as the parent)
	ctxC := logg.CtxWithID(context.Background())
	charlie := logg.New(map[string]interface{}{"c": "C"}).WithID(ctxC)
	charlie.Infof("chicago")
	charlie.Infof("chewbacca")
}

func ExampleNew() {
	// Create a logger without any data fields.
	logger := logg.New(nil)

	// do stuff ...

	logger.Infof("no data fields here")
}

func ExampleNew_multipleSinks() {
	// This logger will emit events to multiple destinations.
	var file, socket io.Writer

	dataFields := map[string]interface{}{
		"bravo":   true,
		"delta":   234 * time.Millisecond,
		"foxtrot": float64(1.23),
		"india":   10,
	}
	logger := logg.New(dataFields, file, socket)

	// do stuff ...

	logger.Infof("hello")
	logger.Infof("world")
}

func ExampleNew_withData() {
	// Initialize the logger with data fields.
	logger := logg.New(map[string]interface{}{
		"bravo":   true,
		"delta":   234 * time.Millisecond,
		"foxtrot": float64(1.23),
		"india":   10,
	})

	// do stuff ...

	logger.Infof("hello")
	logger.Infof("world")
}

func ExampleNew_withID() {
	// Set up a logger a tracing ID.
	logger := logg.New(nil).WithID(context.Background())

	// do stuff ...

	logger.Infof("hello")
	logger.Infof("world")
}
