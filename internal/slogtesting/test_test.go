package slogtesting_test

import (
	"log/slog"
	"testing"
	"time"

	st "github.com/rafaelespinoza/logg/internal/slogtesting"
)

func TestTest(t *testing.T) {
	attrs := []slog.Attr{slog.String("foo", "bar")}
	checks := []st.Test{
		st.TestHasKey("foo"),
		st.TestMissingKey("bar"),
		st.TestHasAttr(slog.String("foo", "bar")),
	}
	for _, check := range checks {
		check(t, attrs)
	}

	groupAttrs := []slog.Attr{
		slog.Time(slog.TimeKey, time.Now()),
		slog.String(slog.LevelKey, slog.LevelInfo.String()),
		slog.String(slog.MessageKey, "msg"),
		slog.String("a", "b"),
		{
			Key: "G",
			Value: slog.GroupValue(
				slog.String("c", "d"),
				slog.Attr{
					Key:   "H",
					Value: slog.GroupValue(slog.String("e", "f")),
				},
			),
		},
	}
	groupChecks := []st.Test{
		st.TestHasKey(slog.TimeKey),
		st.TestHasAttr(slog.String(slog.LevelKey, slog.LevelInfo.String())),
		st.TestHasAttr(slog.String(slog.MessageKey, "msg")),
		st.TestInGroup("G",
			st.TestHasAttr(slog.String("c", "d")),
			st.TestInGroup("H",
				st.TestHasAttr(slog.String("e", "f")),
			),
		),
	}
	for _, check := range groupChecks {
		check(t, groupAttrs)
	}
}

func TestTestFailureExamples(t *testing.T) {
	t.Skip("these are examples of failing tests")

	attrs := []slog.Attr{
		slog.String("bar", "foo"),
		slog.String("g", "golf"),

		slog.String("duplicate_key", "d"),
		slog.String("duplicate_key", "d"),

		{Key: "dupe_group", Value: slog.GroupValue(slog.String("d", "g"))},
		{Key: "dupe_group", Value: slog.GroupValue(slog.String("d", "g"))},

		{Key: "group_with_dupes", Value: slog.GroupValue(slog.String("d", "g"), slog.String("d", "g"))},
	}

	tests := []struct {
		name  string
		check st.Test
	}{
		{name: "has key", check: st.TestHasKey("foo")},
		{name: "missing key", check: st.TestMissingKey("bar")},
		{name: "has attr - key wrong", check: st.TestHasAttr(slog.String("foo", "bar"))},
		{name: "has attr - val wrong", check: st.TestHasAttr(slog.String("bar", "food"))},
		{name: "has attr - enforces 1 attribute with key", check: st.TestHasAttr(slog.String("duplicate_key", "d"))},
		{name: "group - attr not found", check: st.TestInGroup("H", st.TestHasKey("h"))},
		{name: "group - attr found, not a group", check: st.TestInGroup("g", st.TestHasKey("golf"))},
		{name: "group - enforces 1 attribute with key", check: st.TestInGroup("dupe_group", st.TestHasKey("d"))},
		{name: "group - enforces 1 attribute with key in a group", check: st.TestInGroup("group_with_dupes", st.TestHasAttr(slog.String("d", "g")))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.check(t, attrs)
		})
	}
}
