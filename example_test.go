package logg_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rafaelespinoza/logg"
)

// Configure the application logger with some preset fields and a logging destination.
func ExampleConfigure() {
	logg.Configure(
		os.Stderr,
		map[string]string{"branch_name": "main", "build_time": "20060102T150415", "commit_hash": "deadbeef"},
	)
}

// Write to standard error as well as some file.
func ExampleConfigure_multipleSinks() {
	var file io.Writer

	logg.Configure(
		os.Stderr,
		map[string]string{"branch_name": "main", "build_time": "20060102T150415", "commit_hash": "1337d00d"},
		file,
	)
}

// This example shows how the first usage of logg, that is not Configure,
// may unintentionally set up your logger.
func ExampleConfigure_possiblyUnintendedConfiguration() {
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

// Set an ID value and retrieve it. The tracing ID value is what you tell it.
// Derive a new context based on the previous context and set another ID.
func ExampleSetID() {
	const tracingID = "user-generated-random-value"
	ctxA := logg.SetID(context.Background(), tracingID)
	gotID, found := logg.GetID(ctxA)
	fmt.Printf("%q, %t\n", gotID, found)

	// Create a new context derived from the first context.
	const differentTracingID = "different-" + tracingID
	ctxB := logg.SetID(ctxA, differentTracingID)
	gotID, found = logg.GetID(ctxB)
	fmt.Printf("%q, %t\n", gotID, found)
	// Output:
	// "user-generated-random-value", true
	// "different-user-generated-random-value", true
}

// Set an ID value and then retrieve it. Retrieving from contexts without a set
// ID outputs zero values.
func ExampleGetID() {
	const tracingID = "user-generated-random-value"
	ctx := logg.SetID(context.Background(), tracingID)

	gotID, found := logg.GetID(ctx)
	fmt.Printf("%q, %t\n", gotID, found)

	// Demonstrate zero value.
	gotID, found = logg.GetID(context.Background())
	fmt.Printf("%q, %t\n", gotID, found)
	// Output:
	// "user-generated-random-value", true
	// "", false
}

// Create a logger without any data fields.
func ExampleNew_noFields() {
	logger := logg.New(nil)

	// do stuff ...

	logger.Infof("no data fields here")
}

// This logger will emit events to multiple destinations.
func ExampleNew_multipleSinks() {
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

// Initialize the logger with data fields.
func ExampleNew_fields() {
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

// Set up a logger a tracing ID. The tracing ID can be any string you want.
func ExampleNew_withID() {
	ctx := logg.SetID(context.Background(), "logger_id")
	logger := logg.New(nil).WithID(ctx)

	// do stuff ...

	logger.Infof("hello")
	logger.Infof("world")
}

// Demonstrate tracing ID behavior.
func ExampleEmitter_withID() {
	// Create a new context to ensure the same tracing ID on each subsequent
	// tracing event.
	ctxA := logg.SetID(context.Background(), "AAA")
	alfa := logg.New(map[string]interface{}{"a": "A"}).WithID(ctxA)
	// These events have a tracing ID AAA.
	alfa.Infof("altoona")
	alfa.Infof("alice in wonderland")

	// Use the same context in another Emitter.
	bravo := logg.New(map[string]interface{}{"B": "B"}).WithID(ctxA)
	// These events have a tracing ID AAA.
	bravo.Infof("brazil")
	bravo.Infof("bilbo baggins")

	// Deriving a new context with its own tracing ID and passing it to another
	// Emitter yields the tracing ID from the derived context.
	anotherCtx := logg.SetID(ctxA, "CCC")
	charlie := logg.New(map[string]interface{}{"C": "C"}).WithID(anotherCtx)
	// These events have a tracing ID CCC.
	charlie.Infof("chicago")
	charlie.Infof("chewbacca")
}
