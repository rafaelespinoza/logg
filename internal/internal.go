// Package internal is some extra support for log/slog and testing the logg
// package in this module.
package internal

import "log/slog"

// SlogGroupAttrs is a [slog.GroupAttrs] polyfill/shim for golang versions < v1.25.
func SlogGroupAttrs(key string, attrs ...slog.Attr) slog.Attr {
	return slog.Attr{Key: key, Value: slog.GroupValue(attrs...)}
}
