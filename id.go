package logg

import (
	"context"

	"github.com/rs/zerolog"
)

// getSetID retrieves an existing unique id from ctx or creates one. In either
// case, the output is a new context copied from the input.
func getSetID(ctx context.Context, id string) (outCtx context.Context, outID string) {
	outID, ok := GetID(ctx)
	if !ok {
		outID = id
	}
	outCtx = SetID(ctx, outID)
	return
}

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

func newZerologCtxWithID(ctx context.Context, lgr *zerolog.Logger, id string) *zerolog.Context {
	next, nextID := getSetID(ctx, id)
	next = lgr.WithContext(next)
	ztx := zerolog.Ctx(next).With().Str("x_trace_id", nextID)
	return &ztx
}
