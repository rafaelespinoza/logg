package slogtesting_test

import (
	"fmt"
	"log/slog"

	st "github.com/rafaelespinoza/logg/slogtesting"
)

func Example() {
	// Each logging record is collected here.
	var records []slog.Record
	capture := func(r slog.Record) error { records = append(records, r); return nil }

	// Create handler and logger.
	opts := st.AttrHandlerOptions{CaptureRecord: capture}
	handler := st.NewAttrHandler(&opts)
	logger := slog.New(handler)

	// Accumulate some data, output a record at the INFO level.
	logger.With("a", "b").WithGroup("G").With("c", "d").WithGroup("H").Info("msg", "e", "f")

	// Collect 1 record for each output invocation that matches with the
	// handler's output level.
	attrs := st.GetRecordAttrs(records[0])

	// Run these tests.
	checks := []struct {
		check st.Check
		okMsg string
	}{
		{
			check: st.HasKey(slog.TimeKey),
			okMsg: "found key " + slog.TimeKey,
		},
		{
			check: st.HasKey(slog.LevelKey),
			okMsg: "found key " + slog.LevelKey,
		},
		{
			check: st.HasAttr(slog.String(slog.MessageKey, "msg")),
			okMsg: "found attribute with key " + slog.MessageKey,
		},
		{
			check: st.HasAttr(slog.String("a", "b")),
			okMsg: "found attribute with key a",
		},
		{
			check: st.InGroup("G", st.HasAttr(slog.String("c", "d"))),
			okMsg: "found group G and attribute with key c",
		},
		{
			check: st.InGroup("G", st.InGroup("H", st.HasAttr(slog.String("e", "f")))),
			okMsg: "found group G, another group H and attribute with key c",
		},
		{
			check: st.MissingKey("z"),
			okMsg: "did not find attribute with key z",
		},
	}

	for _, ex := range checks {
		err := ex.check(attrs)
		if err != nil {
			fmt.Printf("unexpected error %v\n", err)
		} else {
			fmt.Println(ex.okMsg)
		}
	}
	// Output:
	// found key time
	// found key level
	// found attribute with key msg
	// found attribute with key a
	// found group G and attribute with key c
	// found group G, another group H and attribute with key c
	// did not find attribute with key z
}
