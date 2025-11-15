package slogtesting_test

import (
	"log/slog"
	"testing"
	"time"

	st "github.com/rafaelespinoza/logg/slogtesting"
)

func TestCheckNoGroups(t *testing.T) {
	attrs := []slog.Attr{slog.String("foo", "bar")}
	checks := []st.Check{
		st.HasKey("foo"),
		st.MissingKey("bar"),
		st.HasAttr(slog.String("foo", "bar")),
	}
	for _, check := range checks {
		err := check(attrs)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCheckGroups(t *testing.T) {
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
	groupChecks := []st.Check{
		st.HasKey(slog.TimeKey),
		st.HasAttr(slog.String(slog.LevelKey, slog.LevelInfo.String())),
		st.HasAttr(slog.String(slog.MessageKey, "msg")),
		st.InGroup("G",
			st.HasAttr(slog.String("c", "d")),
			st.InGroup("H",
				st.HasAttr(slog.String("e", "f")),
			),
		),
	}
	for _, check := range groupChecks {
		err := check(groupAttrs)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCheckFailures(t *testing.T) {
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
		check st.Check
	}{
		{name: "has key", check: st.HasKey("foo")},
		{name: "missing key", check: st.MissingKey("bar")},
		{name: "has attr - key wrong", check: st.HasAttr(slog.String("foo", "bar"))},
		{name: "has attr - val wrong", check: st.HasAttr(slog.String("bar", "food"))},
		{name: "has attr - enforces 1 attribute with key", check: st.HasAttr(slog.String("duplicate_key", "d"))},
		{name: "group - attr not found", check: st.InGroup("H", st.HasKey("h"))},
		{name: "group - attr found, not a group", check: st.InGroup("g", st.HasKey("golf"))},
		{name: "group - enforces 1 attribute with key", check: st.InGroup("dupe_group", st.HasKey("d"))},
		{name: "group - enforces 1 attribute with key in a group", check: st.InGroup("group_with_dupes", st.HasAttr(slog.String("d", "g")))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.check(attrs)
			if err == nil {
				t.Error("expected an error but got nil")
			}
		})
	}
}
