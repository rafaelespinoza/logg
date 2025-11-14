package slogtesting_test

import (
	"fmt"
	"log/slog"
	"time"

	st "github.com/rafaelespinoza/logg/slogtesting"
)

func ExampleCheck() {
	attrs := []slog.Attr{
		slog.Time(slog.TimeKey, time.Now()),
		slog.String(slog.LevelKey, slog.LevelInfo.String()),
		slog.String(slog.MessageKey, "Hello, World!"),
	}
	checks := []st.Check{
		st.CheckHasKey(slog.TimeKey),
		st.CheckHasKey("foo"),
		st.CheckMissingKey("foo"),
		st.CheckMissingKey(slog.LevelKey),
		st.CheckHasAttr(slog.String(slog.MessageKey, "Hello, World!")),
		st.CheckHasAttr(slog.String("foo", "bar")),
	}

	for _, check := range checks {
		err := check(attrs)
		fmt.Println(err)
	}
	// Output:
	// <nil>
	// did not find expected key foo
	// <nil>
	// unexpected key level
	// <nil>
	// looking for attr with key foo: unexpected number of matches; got 0, expected 1
}

func ExampleCheckInGroup() {
	attrs := []slog.Attr{
		{Key: "G", Value: slog.GroupValue(slog.String("a", "b"))},
	}

	checkA := st.CheckInGroup("G", st.CheckHasAttr(slog.String("a", "b")))
	err := checkA(attrs)
	fmt.Println(err)

	checkB := st.CheckInGroup("G", st.CheckHasKey("b"))
	err = checkB(attrs)
	fmt.Println(err)
	// Output:
	// <nil>
	// did not find expected key b
}
