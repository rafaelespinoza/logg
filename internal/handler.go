package internal

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
)

type AttrHandler struct {
	opts AttrHandlerOptions
	mtx  *sync.Mutex
	// groupsOrAttrs is a single list of either slog.Groups or []slog.Attr,
	// accumulated via calls to the WithAttrs, WithGroup methods. This list is
	// used when it's time to log something via the Handle method.
	groupsOrAttrs []groupOrAttrs
}

const logPrefix = "logg/internal: "

type AttrHandlerOptions struct {
	slog.HandlerOptions
	CaptureRecord func(r slog.Record) error
}

func NewAttrHandler(opts *AttrHandlerOptions) *AttrHandler {
	if opts == nil {
		opts = &AttrHandlerOptions{}
	}
	return &AttrHandler{
		opts:          *opts,
		mtx:           &sync.Mutex{},
		groupsOrAttrs: make([]groupOrAttrs, 0, 4),
	}
}

func (h *AttrHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	level := slog.LevelInfo
	if h.opts.Level != nil {
		level = h.opts.Level.Level()
	}
	enabled := lvl >= level
	return enabled
}

func (h *AttrHandler) Handle(ctx context.Context, rec slog.Record) (err error) {
	recordAttrs := h.buildRecordAttrs(rec)
	cloned := slog.NewRecord(rec.Time, rec.Level, rec.Message, rec.PC)
	cloned.AddAttrs(recordAttrs...)

	slog.Debug(logPrefix+"called .Handle",
		slog.Int("num_attrs_on_input_record", rec.NumAttrs()),
		slog.Int("num_attrs_on_processed_record", cloned.NumAttrs()),
	)

	h.mtx.Lock()
	defer h.mtx.Unlock()
	if capture := h.opts.CaptureRecord; capture != nil {
		err = capture(cloned)
	}
	return
}

func (h *AttrHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var out AttrHandler
	defer func() {
		slog.Debug(logPrefix+"called .WithAttrs",
			slog.Int("num_input_attrs", len(attrs)), slog.Any("attrs", attrs),
			slog.Any("num_goas_before", len(h.groupsOrAttrs)), slog.Any("num_goas_after", len(out.groupsOrAttrs)),
		)
	}()
	if len(attrs) == 0 {
		return h
	}

	out = *h
	out.groupsOrAttrs = copyAppend(h.groupsOrAttrs, groupOrAttrs{Attrs: attrs})

	return &out
}

func (h *AttrHandler) WithGroup(name string) slog.Handler {
	var out AttrHandler
	defer func() {
		slog.Debug(logPrefix+"called .WithGroup",
			slog.String("name", name),
			slog.Any("num_goas_before", len(h.groupsOrAttrs)), slog.Any("num_goas_after", len(out.groupsOrAttrs)),
		)
	}()
	if name == "" {
		return h
	}

	out = *h
	out.groupsOrAttrs = copyAppend(h.groupsOrAttrs, groupOrAttrs{GroupName: name})

	return &out
}

func copyAppend[T any](originals []T, tail ...T) (out []T) {
	nextLen := len(originals) + len(tail)
	out = make([]T, len(originals), nextLen)
	copy(out, originals)
	out = append(out, tail...)
	return
}

// groupOrAttrs helps the slog.Handler maintain internal state of groups added
// via slog.Handler.WithGroup, or attributes added via slog.Handler.WithAttrs.
// The GroupName field is non-empty when created through WithGroup. The Attrs
// field has a non-zero length when created through WithAttrs. Field names are
// upper-cased to help us see the state of the fields when using the
// slog.JSONHandler for debugging.
type groupOrAttrs struct {
	GroupName string
	Attrs     []slog.Attr
}

func (h *AttrHandler) buildRecordAttrs(rec slog.Record) (out []slog.Attr) {
	out = make([]slog.Attr, 0, 4+rec.NumAttrs())

	// Place the built-in attributes at the top-level.
	replaceAttr := h.opts.ReplaceAttr

	// From the log/slog.Handler docs:
	// 	- If r.Time is the zero time, ignore the time.
	if !rec.Time.IsZero() {
		timeAttr := slog.Time(slog.TimeKey, rec.Time)
		if replaceAttr != nil {
			timeAttr = replaceAttr([]string{}, timeAttr)
		}
		out = append(out, timeAttr)
	}

	levelAttr := slog.Any(slog.LevelKey, rec.Level)
	if replaceAttr != nil {
		levelAttr = replaceAttr([]string{}, levelAttr)
	}
	out = append(out, levelAttr)

	msgAttr := slog.String(slog.MessageKey, rec.Message)
	if replaceAttr != nil {
		msgAttr = replaceAttr([]string{}, msgAttr)
	}
	out = append(out, msgAttr)

	// Now prepare the non-built-in attributes. Start with groups and attrs that
	// were accumulated on the handler.
	goas := h.groupsOrAttrs
	if rec.NumAttrs() < 1 {
		// If the record has no Attrs, remove groups at the end of the list; they are empty.
		for len(goas) > 0 && goas[len(goas)-1].GroupName != "" {
			goas = goas[:len(goas)-1]
		}
	}

	var groupPath []string
	for _, goa := range goas {
		if goa.GroupName != "" {
			groupPath = append(groupPath, goa.GroupName)
		} else {
			subGroup := applyGroupsToAttrs(groupPath, goa.Attrs)
			if replaceAttr != nil {
				for i, attr := range subGroup {
					subGroup[i] = replaceAttr(groupPath, attr)
				}
			}
			out = mergeAttrs(out, subGroup)
			slog.Debug(logPrefix+"from .buildRecordAttrs in goas loop",
				slog.Any("group_path", groupPath), slog.Any("attrs_added", goa.Attrs),
				slog.Any("grouped_attrs", subGroup), slog.Any("out", out),
			)
		}
	}

	recordAttrs := make([]slog.Attr, 0, rec.NumAttrs())
	rec.Attrs(func(attr slog.Attr) bool {
		// From the log/slog.Handler docs:
		// 	- Attr's values should be resolved.
		attr.Value = attr.Value.Resolve()

		if attr.Equal(slog.Attr{}) {
			// From the log/slog.Handler docs:
			// 	- If an Attr's key and value are both the zero value, ignore the Attr.
			return true
		}

		if attr.Value.Kind() != slog.KindGroup {
			recordAttrs = append(recordAttrs, attr)
		} else {
			group := attr.Value.Group()

			if attr.Key != "" {
				recordAttrs = append(recordAttrs, SlogGroupAttrs(attr.Key, group...))
			} else {
				// From the log/slog.Handler docs:
				// 	- If a group's key is empty, inline the group's Attrs.
				recordAttrs = append(recordAttrs, group...)
			}
		}

		return true
	})

	if replaceAttr != nil {
		for i, attr := range recordAttrs {
			recordAttrs[i] = replaceAttr(groupPath, attr)
		}
	}

	appliedRecordAttrs := applyGroupsToAttrs(groupPath, recordAttrs)
	if replaceAttr != nil {
		for i, attr := range appliedRecordAttrs {
			appliedRecordAttrs[i] = replaceAttr(groupPath, attr)
		}
	}
	if len(groupPath) < 1 {
		out = append(out, appliedRecordAttrs...)
		return
	}

	out = mergeAttrs(out, appliedRecordAttrs)
	slog.Debug(logPrefix+"from .buildRecordAttrs",
		slog.Any("group_path", groupPath), slog.Any("record_attrs", recordAttrs),
		slog.Any("applied_record_attrs", appliedRecordAttrs), slog.Any("out", out),
	)
	return
}

func applyGroupsToAttrs(groups []string, data []slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		return data
	}

	// Start from the innermost group and wrap outward. Iterate backward through
	// the groups slice. For example, if groups is ["level1", "level2"], we
	// first wrap data in a group with key "level2", then wrap that result in a
	// group with key "level1".
	curr := SlogGroupAttrs(groups[len(groups)-1], data...)

	for i := len(groups) - 2; i >= 0; i-- {
		groupName := groups[i]

		// Wrap the current group in another group and move up a level.
		curr = SlogGroupAttrs(groupName, curr)
	}

	return []slog.Attr{curr}
}

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
	// As a special case for when both attributes are groups, merge them
	// together rather than overwrite entirely.
	for _, next := range nextList {
		prev, alreadyExists := attrsBykey[next.Key]
		if alreadyExists && prev.Value.Kind() == slog.KindGroup && next.Value.Kind() == slog.KindGroup {
			merged, err := mergeGroups(prev, next)
			if err == nil {
				attrsBykey[next.Key] = merged
				continue
			}

			// in an error case, do a fallthrough and let the previous value be
			// overridden by the next value.
		}

		attrsBykey[next.Key] = next
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

func mergeGroups(prev, next slog.Attr) (slog.Attr, error) {
	prevKind, nextKind := prev.Value.Kind(), next.Value.Kind()
	if prevKind != slog.KindGroup || nextKind != slog.KindGroup {
		err := fmt.Errorf(
			"%s when attempting to merge groups, both attributes must be of kind Group. prev kind=%s, next kind=%s",
			logPrefix, prevKind.String(), nextKind.String(),
		)
		return slog.Attr{}, err
	}

	prevGroup, nextGroup := prev.Value.Group(), next.Value.Group()
	merged := mergeAttrs(prevGroup, nextGroup)
	out := SlogGroupAttrs(prev.Key, merged...)
	return out, nil
}

// GetRecordAttrs collects each attribute on the record.
func GetRecordAttrs(r slog.Record) []slog.Attr {
	out := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		out = append(out, a)
		return true
	})
	return out
}

// SlogGroupAttrs is a [slog.GroupAttrs] polyfill/shim for golang versions < v1.25.
func SlogGroupAttrs(key string, attrs ...slog.Attr) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.GroupValue(attrs...),
	}
}
