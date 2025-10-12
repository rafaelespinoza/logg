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

Package logg is a thin wrapper around log/slog. The primary goal is to abstract
structured logging for an application while providing a simpler API. It's
opinionated and offers a limited feature set.

The feature set is:

- attribute management, including contextual tracing IDs
- timestamps
- leveled logging (only ERROR and INFO severities)
- emit JSON, TEXT (space separated key=value pairs)
  - pass in a `slog.Handler` for further customization

## Usage

Call the `Setup` function as early as possible in your application. This
initializes a root logger, which functions like a prototype for subsequent
events. Things initialized are the output sinks and an optional "version"
field. The "version" data appears in its own group for each log event.

Use the `Error`, `Info` functions to log at error, info levels respectively.

Tracing IDs are managed with a context API, see the `SetID` and `GetID`
functions. You supply the value. There are many choices in this area, some
examples are:

- [github.com/gofrs/uuid](https://github.com/gofrs/uuid)
- [github.com/google/uuid](https://github.com/google/uuid)
- [github.com/rs/xid](https://github.com/rs/xid)
- [github.com/segmentio/ksuid](https://github.com/segmentio/ksuid)

To add more event-specific fields to a logging entry, call `New` and then call
one of the `Emitter` methods. Call the `Emitter.WithData` method to
independently decorate subsequent events based upon previous events. Use the
`Emitter.WithID` to pass in a context with a tracing ID.

See more in the godoc examples.

## Event shape

When using the `slog.JSONHandler` with default settings, these top-level fields
are present:

- `time`: string, rfc3339 timestamp.
- `level`: string, either `"INFO"`, `"ERROR"`
- `msg`: string, what happened

These top-level fields may or may not be present, depending on configuration and
how the event is emitted:
- `version`: string key value tuples, optional versioning metadata from your
  application. Will only be present when this data is passed in to the `Setup`
  function.
- `error`: string, an error message. Only when the event is emitted with an
  Error level.
- `x_trace_id`: string, a tracing ID. Present when a non-empty value is set on
  the context passed into the `Emitter.WithID` method.
- `data`: map[string]any, event-specific fields. Only present when an
  event emitter is constructed with fields via the `New` function and/or is
  passed something to its `WithData` method.

### Example events

These examples use the default handler in this library, `slog.JSONHandler`.

Info level
```
{
  "time":"2025-09-22T08:59:52.657053724Z",
  "level":"INFO",
  "msg":"TestLogger/empty_sinks",
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

Error level
```
{
  "time":"2025-09-22T08:47:11.856469998Z",
  "level":"ERROR",
  "msg":"database library error",
  "error":"pq: duplicate key value violates unique constraint \"unique_index_on_foos_bar_id\"",
  "data":{
    "code":"23505",
    "message":"duplicate key value violates unique constraint \"unique_index_on_foos_bar_id\"",
    "table":"foos"
  }
}
```

Versioning metadata can be added with the `Setup` function.

```
{
  "time":"2025-09-22T08:59:52.347111271Z",
  "level":"INFO",
  "msg":"TestLogger/empty_sinks",
  "version":{
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
