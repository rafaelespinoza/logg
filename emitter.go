package logg

import (
	"context"
	"log/slog"
	"slices"
)

// An Emitter logs events with preset version info and attributes data. Set a
// tracing ID on a context.Context with the [SetID] function and pass it into
// the [Emitter.WithID] method. Log output methods are Debug, Info, Warn, Error
// which write at the respective logging levels.
//
// # Data attribute management
//
// This section describes how attributes are managed for the output log's "data"
// key. When an Emitter is created via the [New] function it takes slog.Attr as
// input and prepares the attributes to be present in future logging events for
// the created Emitter. The [Emitter.WithData] method combines the input
// attributes with the Emitter's existing attributes and returns a new Emitter
// with the resulting attributes. The resulting list may be subject to
// deduplication by the attribute key. The logic is:
//
//   - If one of the lists is length 0 but the other list is not, then the list
//     with a non-zero length is returned. If both lists are length 0, then a
//     new list of length 0 is returned. If both lists have a non-zero length,
//     then it continues on.
//   - If a key is duplicated within the same list, then the item that appears
//     later in that list has higher precedence.
//   - If a key is found in the list of existing attributes and in the list of
//     input attributes, then the item from the input list has higher precedence
//     than an item with the same key from the existing list.
//   - If a key is unique to either list of attributes, then that attribute will
//     be in the resulting list.
//   - Only top-level keys are considered in each list. When both lists have an
//     attribute with the same key and both are [slog.KindGroup], there is no
//     deep merge.
//   - The returned list is a new slice. Neither input slice is modified.
//
// The log output methods, when passed input attributes and when the Emitter is
// enabled, also apply this logic. A notable difference with the
// [Emitter.WithData] method is that the created attributes are not set to the
// Emitter. Instead, the resulting attributes are sent directly to the log.
type Emitter struct {
	lgr        *slog.Logger
	versioning []slog.Attr
	id         string
	attrs      []slog.Attr
}

// New initializes a logger type and configures it so each event emission
// outputs attrs at the "data" key. If h is nil then it uses a default root
// handler, which is configured via the [Setup] function.
func New(h slog.Handler, dataAttrs ...slog.Attr) *Emitter {
	defaults := rootEmitter()
	if h == nil {
		h = defaults.lgr.Handler()
	}

	lgr := newSlogger(h, defaults.versioning...)
	out := newEmitter(lgr, defaults.versioning, "", dataAttrs...)
	return out
}

// Debug writes msg at the DEBUG level with optional attributes.
func (l *Emitter) Debug(msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelDebug
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, nil, msg, l.id, mergedAttrs...)
}

// Info writes msg at the INFO level with optional attributes.
func (l *Emitter) Info(msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelInfo
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, nil, msg, l.id, mergedAttrs...)
}

// Warn writes msg at the WARN level with optional attributes.
func (l *Emitter) Warn(msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelWarn
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, nil, msg, l.id, mergedAttrs...)
}

// Error writes msg and err to the log at the ERROR level with optional
// attributes. The input err is placed on to the "error" key in the log event.
func (l *Emitter) Error(err error, msg string, attrs ...slog.Attr) {
	ctx := context.Background()
	lvl := slog.LevelError
	if !l.lgr.Enabled(ctx, lvl) {
		return // prevent unnecessary attr merging
	}

	mergedAttrs := mergeAttrs(l.attrs, attrs)
	log(ctx, l.lgr, lvl, err, msg, l.id, mergedAttrs...)
}

// WithID reads the input ctx and prepares the Emitter to write the tracing ID
// at the key, "x_trace_id". Use the [SetID] function to prepare the input
// context. If a tracing ID already existed on the Emitter, then it's replaced.
func (l *Emitter) WithID(ctx context.Context) *Emitter {
	l.id, _ = GetID(ctx)
	return l
}

// WithData creates a new Emitter with added attributes. When there is an
// existing attribute with the same key as an input attribute, then the input
// attribute replaces the existing attribute.
func (l *Emitter) WithData(attrs []slog.Attr) *Emitter {
	versioning := rootEmitter().versioning
	mergedAttrs := mergeAttrs(l.attrs, attrs)

	return newEmitter(l.lgr, versioning, l.id, mergedAttrs...)
}

func newEmitter(lgr *slog.Logger, versioning []slog.Attr, id string, dataAttrs ...slog.Attr) *Emitter {
	return &Emitter{
		lgr:        lgr,
		versioning: versioning,
		id:         id,
		attrs:      dataAttrs,
	}
}

func newSlogger(handler slog.Handler, versioningAttrs ...slog.Attr) *slog.Logger {
	group := slog.GroupAttrs(versionFieldName, versioningAttrs...)
	lgr := slog.New(handler).With(group)
	lgr.Debug(libraryMsgPrefix + "initialized logger")
	return lgr
}

// mergeAttrs combines, deduplicates the input lists into an output list. It
// returns early if 1 or both lists are length 0. Otherwise, it takes 2 passes
// over each list.
//
//  1. The 1st pass upon each list collects and deduplicate items based on the
//     Attr.Key. items appearing later replace items seen earlier. The prevList is
//     visited before the nextList.
//  2. The 2nd pass upon each list, constructs an output list based on the
//     deduplicated items found in the 1st pass. The ordering is based upon the
//     nextList; items unique to the prevList are added at the end.
func mergeAttrs(prevList, nextList []slog.Attr) []slog.Attr {
	if len(prevList) > 0 && len(nextList) < 1 {
		return prevList
	} else if len(prevList) < 1 && len(nextList) > 0 {
		return nextList
	} else if len(prevList) < 1 && len(nextList) < 1 {
		return make([]slog.Attr, 0)
	}

	maxLength := max(len(prevList), len(nextList))

	// 2 passes over each list.
	// 1st pass: collect items to add.
	// 2nd pass: build output while preserving original order of keys.

	//
	// 1st pass
	//
	attrsBykey := make(map[string]slog.Attr, maxLength)

	// Start with items from prevList. Items that are later in prevList take
	// precedence over items with the same key that appear earlier in prevList.
	for _, attr := range prevList {
		attrsBykey[attr.Key] = attr
	}

	// Then look through nextList. Same idea as prevList: allow later items with
	// same key to take precedence over earlier items. But also, any item in
	// nextList with the same key as an item in prevList will take precedence.
	for _, attr := range nextList {
		attrsBykey[attr.Key] = attr
	}

	//
	// 2nd pass
	//
	addedKeys := make(map[string]struct{}, maxLength)
	out := make([]slog.Attr, 0, maxLength)
	var found bool

	// Look at items collected from nextList, and add them to output unless
	// we've already seen the value.
	for _, attr := range nextList {
		if _, found = addedKeys[attr.Key]; !found {
			out = append(out, attrsBykey[attr.Key])
			addedKeys[attr.Key] = struct{}{} // prevent further re-adding to output list
		}
	}

	// Now look at items from prevList, and only add the item if its key is
	// unique to prevList.
	for _, attr := range prevList {
		if _, found = addedKeys[attr.Key]; !found {
			out = append(out, attrsBykey[attr.Key])
			addedKeys[attr.Key] = struct{}{} // prevent further re-adding to output list
		}
	}

	return slices.Clip(out)
}
