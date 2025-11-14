package slogtesting

import (
	"log/slog"
	"testing"
)

// A Test is a general-purpose test on slog attributes. It's based off the
// testing/slogtest package.
type Test func(*testing.T, []slog.Attr)

// TestHasKey makes a Test for the presence of an attribute with a key.
func TestHasKey(key string) Test {
	ch := CheckHasKey(key)
	return checkToTest(ch)
}

// TestMissingKey makes a Test for the absence of an attribute with a key.
func TestMissingKey(key string) Test {
	ch := CheckMissingKey(key)
	return checkToTest(ch)
}

// TestHasAttr makes a Test for the presence of an attribute with the wanted key
// and value.
func TestHasAttr(want slog.Attr) Test {
	ch := CheckHasAttr(want)
	return checkToTest(ch)
}

// TestInGroup makes a Test for a Test in a group with a matching name.
func TestInGroup(name string, a Test, as ...Test) Test {
	return func(t *testing.T, attrs []slog.Attr) {
		t.Helper()

		matchKey := makeKeyMatcher(name)
		got, err := collectNMatchingAttrs(attrs, 1, matchKey)
		if err != nil {
			t.Errorf("looking for group attr with name %s: %v", name, err)
			return
		}

		kind := got[0].Value.Kind()
		if kind != slog.KindGroup {
			t.Fatalf("wrong kind (%s) for item with key %s, expected %s", kind, name, slog.KindGroup.String())
		}

		groupVals := got[0].Value.Group()
		a(t, groupVals)
		for _, check := range as {
			check(t, groupVals)
		}
	}
}

func checkToTest(ch Check) Test {
	return func(t *testing.T, attrs []slog.Attr) {
		t.Helper()

		err := ch(attrs)
		if err != nil {
			t.Error(err)
		}
	}
}
