// Package internal is some extra support for log/slog and testing the logg
// package in this module.
package internal

import "log/slog"

const logPrefix = "logg/internal: "

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
	return slog.Attr{Key: key, Value: slog.GroupValue(attrs...)}
}
