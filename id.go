package logg

import "context"

type traceIDKey struct{}

// SetID puts val on the context for use as a tracing ID. If a tracing ID
// already existed on the input ctx, then it's replaced. Use the [GetID]
// function to retrieve the tracing ID.
func SetID(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, val)
}

// GetID fetches a tracing ID value from context if it's found. If it's not
// found, then the tracing ID is empty. Use the [SetID] function to place a
// tracing ID on to the context.
func GetID(ctx context.Context) (id string, found bool) {
	id, found = ctx.Value(traceIDKey{}).(string)
	return
}
