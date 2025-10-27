```
 __        ______     _______   _______
|  |      /  __  \   /  _____| /  _____|
|  |     |  |  |  | |  |  __  |  |  __
|  |     |  |  |  | |  | |_ | |  | |_ |
|  `----.|  `--'  | |  |__| | |  |__| |
|_______| \______/   \______|  \______|
```

[![](https://github.com/rafaelespinoza/logg/workflows/build/badge.svg)](https://github.com/rafaelespinoza/logg/actions)
[![](https://pkg.go.dev/badge/github.com/rafaelespinoza/logg)](https://pkg.go.dev/github.com/rafaelespinoza/logg)
[![codecov](https://codecov.io/gh/rafaelespinoza/logg/branch/main/graph/badge.svg?token=GFUSTO55PY)](https://codecov.io/gh/rafaelespinoza/logg)

Package logg is a thin wrapper around log/slog. The primary goal is to leverage
all the things offered by that package, but also make it easier to separate
application metadata from event-specific data. It's opinionated and offers a
limited feature set.

The feature set is:

- It's `log/slog` with light attribute management. After package setup,
  attributes and groups added to a logger via `slog.Logger.With` or
  `slog.Logger.WithGroup` are placed under a pre-built group. Attributes passed
  to a log output method are also placed under this group.
- Pass in a `slog.Handler` for further customization.

## Usage

Call the `SetDefaults` function as early as possible in your application. This
initializes a root logger, which functions like a prototype for subsequent
events. Things initialized are the output sinks and an optional
`"application_metadata"` field. The data appears in its own group for each log
event.

To add more event-specific fields to a logging entry, call `New`. This function
creates a `slog.Logger` with an optional trace ID and optional data attributes.
Attributes, including group attributes, added through the logger methods, would
be applied to one pre-built group attribute.

See more in the godoc examples.

## Event shape

When using the `slog.JSONHandler` with default settings, these top-level fields
are present:

- `time`: string, rfc3339 timestamp.
- `level`: string, sometimes this is called "severity". One of: `"DEBUG"`,
  `"INFO"`,`"WARN"`, `"ERROR"`.
- `msg`: string, what happened.

These top-level fields may or may not be present, depending on configuration and
how the event is emitted:
- `application_metadata`: map[string]any, optional versioning metadata from
  your application. Will only be present when this data is passed in to the
  `SetDefaults` function.
- `trace_id`: string, a tracing ID. Present when a non-empty value is passed
  into the `New` function.
- `data`: map[string]any, other event-specific fields.

### Example events

These examples use the `slog.Handler` implementation `slog.JSONHandler`.

Info level
```
{
  "time":"2025-09-22T08:59:52.657053724Z",
  "level":"INFO",
  "msg":"TestLogger",
  "data":{
    "alfa": "anything",
    "bravo": {
      "bool": true,
      "duration_ns": 1234,
      "float": 1.23,
      "int": 10,
      "string": "nevada"
    }
  }
}
```

Versioning metadata can be added with the `SetDefaults` function.
```
{
  "time":"2025-09-22T08:59:52.347111271Z",
  "level":"INFO",
  "msg":"TestLogger",
  "application_metadata":{
    "branch_name":"main",
    "go_version":"v1.25",
    "commit_hash":"deadbeef"
  },
  "data":{
    "alfa": "anything",
    "bravo": {
      "bool": true,
      "duration_ns": 1234,
      "float": 1.23,
      "int": 10,
      "string": "nevada"
    }
  }
}
```
