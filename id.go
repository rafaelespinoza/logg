package logg

import (
	"context"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// CtxWithID returns a new context with an ID. If the ID already existed in the
// context, then the new context has the same ID as before.
func CtxWithID(ctx context.Context) context.Context {
	out, _ := getSetID(ctx)
	return out
}

// getSetID retrieves an existing unique id from ctx or creates one. In either
// case, the output is a new context copied from the input.
func getSetID(ctx context.Context) (out context.Context, id string) {
	xID, ok := hlog.IDFromCtx(ctx)
	if !ok {
		xID = xid.New()
	}
	out = hlog.CtxWithID(ctx, xID)
	id = xID.String()
	return
}

func newZerologCtxWithID(ctx context.Context, lgr *zerolog.Logger) *zerolog.Context {
	next, id := getSetID(ctx)
	next = lgr.WithContext(next)
	ztx := zerolog.Ctx(next).With().Str("x_trace_id", id)
	return &ztx
}
