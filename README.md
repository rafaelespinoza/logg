```
 __        ______     _______   _______
|  |      /  __  \   /  _____| /  _____|
|  |     |  |  |  | |  |  __  |  |  __
|  |     |  |  |  | |  | |_ | |  | |_ |
|  `----.|  `--'  | |  |__| | |  |__| |
|_______| \______/   \______|  \______|
```

Package logg is merely a wrapper around github.com/rs/zerolog. The primary goal
is to abstract structured logging for an application while providing a simpler
API. It's rather opinionated, and offers a limited feature set.

The feature set is:

- provide timestamps
- unique event ids
- logging levels (just error and info)
- emit JSON

## Usage

Call the `Configure` function as early as possible in your application. This
initializes a "root" logger, which functions like a prototype for all subsequent
events. Things initialized are the output sinks and an optional "version" field.
The "version" is just some application versioning metadata, which may be useful
if you want to know something about your application's source code.

See the godoc examples for details.

Use the `Errorf`, `Infof` functions to log at error, info levels respectively.
To add more event-specific fields to a logging entry, create an event with
`NewEvent` and call one of the `Emitter` methods.

## Event shape

These top-level fields are always present:

- `level`: string, either `"info"`, `"error"`
- `time`: string, an rfc3339 timestamp of emission in system's timezone.
- `message`: string, what happened

These top-level fields may or may not be present, depending on configuration and
how the event is emitted:
- `error`: string, an error message. only when the event is emitted with an
  Error level.
- `version`: map[string]string, optional versioning metadata from your
  application. will only be present when this data is passed in to the
  `Configure` function.

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
