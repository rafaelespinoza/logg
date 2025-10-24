package internal_test

import (
	"encoding/json"
	"io"
	"log/slog"
)

func makeJSONRecordCapturer(w io.Writer) func(r slog.Record) error {
	return func(r slog.Record) error {
		attrs := getRecordAttrs(r)
		mappedAttrs := mapAttrs(attrs)
		return json.NewEncoder(w).Encode(mappedAttrs)
	}
}

func getRecordAttrs(r slog.Record) []slog.Attr {
	out := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		out = append(out, a)
		return true
	})
	return out
}

func mapAttrs(attrs []slog.Attr) map[string]any {
	out := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		if attr.Equal(slog.Attr{}) {
			continue // discard empty attributes
		}

		attr.Value = attr.Value.Resolve()

		if attr.Value.Kind() != slog.KindGroup {
			out[attr.Key] = slogValueToAny(attr.Value)
		} else {
			group := attr.Value.Group()
			if len(group) < 1 {
				continue // discard groups without any attributes
			}

			if attr.Key != "" {
				out[attr.Key] = mapAttrs(group)
			} else {
				// inline the attributes
				mergeMaps(out, mapAttrs(group))
			}
		}
	}

	return out
}

func slogValueToAny(val slog.Value) (out any) {
	val = val.Resolve()
	switch val.Kind() {
	case slog.KindAny:
		out = val.Any()
	case slog.KindBool:
		out = val.Bool()
	case slog.KindDuration:
		out = val.Duration()
	case slog.KindFloat64:
		out = val.Float64()
	case slog.KindInt64:
		out = val.Int64()
	case slog.KindString:
		out = val.String()
	case slog.KindTime:
		out = val.Time()
	case slog.KindUint64:
		out = val.Uint64()
	case slog.KindGroup:
		out = mapAttrs(val.Group())
	case slog.KindLogValuer:
		out = val.LogValuer().LogValue()
	default:
		out = val.Any()
	}
	return
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if subSrc, srcIsMap := v.(map[string]any); !srcIsMap {
			dst[k] = v
		} else {
			subDst, dstIsMap := dst[k].(map[string]any)
			if !dstIsMap {
				dst[k] = subSrc
			} else {
				mergeMaps(subDst, subSrc)
			}
		}
	}
}
