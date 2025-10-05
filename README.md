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

Package logg is merely a wrapper around github.com/rs/zerolog. The primary goal
is to abstract structured logging for an application while providing a simpler
API. It's rather opinionated, and offers a limited feature set.

The feature set is:

- provide timestamps
- tracing ids (user-provided)
- leveled logging (only ERROR and INFO severities)
- emit JSON

## Usage

Call the `Configure` function as early as possible in your application. This
initializes a "root" logger, which functions like a prototype for all subsequent
events. Things initialized are the output sinks and an optional "version" field.
The "version" is just some application versioning metadata, which may be useful
if you want to know something about your application's source code.

Use the `Errorf`, `Infof` functions to log at error, info levels respectively.

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

These top-level fields are always present:

- `level`: string, either `"info"`, `"error"`
- `time`: string, an rfc3339 timestamp of emission in system's timezone.
- `message`: string, what happened

These top-level fields may or may not be present, depending on configuration and
how the event is emitted:
- `error`: string, an error message. Only when the event is emitted with an
  Error level.
- `version`: map[string]string, optional versioning metadata from your
  application. Will only be present when this data is passed in to the
  `Configure` function.
- `data`: map[string]interface{}, event-specific fields. Only present when an
  event emitter is constructed with fields via the `New` function and/or is
  passed something to its `WithData` method.

### Example events

Info level
```
{
  "level":"info",
  "time":"2021-09-22T08:59:52-07:00",
  "message":"TestLogger/empty_sinks",
  "data":{
    "alfa": "anything",
    "bravo": {
      "bool": true,
      "duration": 1234,
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
  "level":"error",
  "error":"pq: duplicate key value violates unique constraint \"unique_index_on_foos_bar_id\"",
  "time":"2021-09-22T08:47:11-07:00",
  "message":"database library error",
  "data":{
    "code":"23505",
    "message":"duplicate key value violates unique constraint \"unique_index_on_foos_bar_id\"",
    "table":"foos"
  }
}
```

Versioning metadata can be added with the `Configure` function, which must be
invoked before the first invocation of any library function.

```
{
  "level":"info",
  "time":"2021-09-22T08:59:52-07:00",
  "message":"TestLogger/empty_sinks",
  "version":{
    "branch_name":"main",
    "go_version":"v1.17",
    "commit_hash":"deadbeef"
  },
  "data":{
    "alfa": "anything",
    "bravo": {
      "bool": true,
      "duration": 1234,
      "float": 1.23,
      "int": 10,
      "string": "nevada"
    }
  }
}
```
