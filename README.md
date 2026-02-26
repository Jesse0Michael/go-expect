# go-expect

[![Build](https://github.com/Jesse0Michael/go-expect/workflows/Build/badge.svg)](https://github.com/Jesse0Michael/go-expect/actions?query=branch%3Amain) [![Coverage Status](https://coveralls.io/repos/github/Jesse0Michael/go-expect/badge.svg?branch=main)](https://coveralls.io/github/Jesse0Michael/go-expect?branch=main)

API end-to-end testing in Go. Define multi-step scenarios that share state across requests — using either a fluent Go builder or YAML/JSON files — and run them with `go test`.

```go
import "github.com/jesse0michael/go-expect/pkg/expect"
```

---

## Concepts

| Type | Role |
|------|------|
| `Suite` | Collection of scenarios sharing a set of named connections |
| `Scenario` | Ordered sequence of steps; variables flow from one step to the next |
| `Step` | Single request + assertion pair |
| `Connection` | Named target (`HTTP` or `GRPC`); the first registered becomes the default |
| `VarStore` | `map[string]any` shared across steps — populated by `Save`, consumed via `{key}` interpolation |

---

## Go API

### HTTP

```go
srv := httptest.NewServer(handler)
defer srv.Close()

suite := expect.NewSuite().
    WithConnections(expect.HTTP("api", srv.URL)).
    WithScenarios(
        expect.NewScenario("user lifecycle").
            AddStep(
                expect.POST("/users").
                    WithJSON(map[string]any{"name": "alice"}).
                    ExpectStatus(201).
                    Save("id", "user_id"),
            ).
            AddStep(
                expect.GET("/users/{user_id}").
                    ExpectStatus(200).
                    ExpectBody(map[string]any{"name": "alice"}),
            ).
            AddStep(
                expect.DELETE("/users/{user_id}").
                    ExpectStatus(204),
            ),
    )

expect.NewTestSuite(suite).Run(t)
```

**Builder methods:**

| Method | Notes |
|--------|-------|
| `GET / POST / PUT / PATCH / DELETE(path)` | Shorthand constructors |
| `HTTPStep(method, path)` | Arbitrary method |
| `WithConnection(name)` | Override the default connection for this step |
| `WithHeader(key, value)` | Request header |
| `WithQuery(key, value)` | URL query parameter |
| `WithBody([]byte)` | Raw request body |
| `WithJSON(v any)` | Marshal to JSON; sets `Content-Type: application/json` |
| `ExpectStatus(code int)` | Exact status code |
| `ExpectHeader(key, value)` | Response header assertion |
| `ExpectBody(v any)` | Partial JSON match (or exact bytes/string) |
| `Save(field, as)` | Extract a top-level response field into a variable |

### gRPC

```go
// Typed — compiled proto stubs
expect.GRPCCall("svc", "/pkg.MyService/GetUser", &mypb.GetUserRequest{Id: 42}).
    ExpectGRPCCode("OK").
    ExpectGRPCBody(map[string]any{"name": "alice"})

// Raw — no stubs required
expect.GRPCRawCall("svc", "/pkg.MyService/GetUser", []byte(`{"id":42}`)).
    ExpectGRPCCode("OK").
    SaveGRPC("name", "user_name")
```

| Method | Notes |
|--------|-------|
| `GRPCCall(conn, fullMethod, proto.Message)` | Typed invocation via compiled stubs |
| `GRPCRawCall(conn, fullMethod, []byte)` | Raw JSON invocation; no stubs needed |
| `ExpectGRPCCode(code string)` | gRPC status code name: `"OK"`, `"NOT_FOUND"`, etc. |
| `ExpectGRPCBody(v any)` | Partial JSON match against response |
| `SaveGRPC(field, as)` | Extract a field from JSON response into a variable |

### Hooks

```go
expect.NewScenario("seeded test").
    Before(func() error {
        return db.Seed(testData)
    }).
    After(func() error {
        return db.Reset()
    }).
    AddStep(...)
```

`Before` functions gate step execution — if any `Before` fails, steps are skipped. `After` functions always run regardless.

---

## YAML / JSON

Load test suites from files with `LoadFile`, `LoadDir`, or `LoadFS` (for `//go:embed`).

```go
// Single file
suite, err := expect.LoadFile("testdata/expect.yaml")
suite.WithConnections(expect.HTTP("api", srv.URL)) // override connection URL at runtime

// Directory — all *.yaml, *.yml, *.json files merged into one Suite
suite, err := expect.LoadDir("testdata/")

// Embedded FS
//go:embed testdata
var testFS embed.FS
suite, err := expect.LoadFS(testFS)
```

### Schema

```yaml
connections:
  - name: api          # referenced by steps; first entry is the default
    type: http         # "http", "https", or "grpc"
    url: http://localhost:8080

scenarios:
  - name: counter flow
    steps:
      - request:
          connection: api   # omit to use the default connection
          method: POST
          endpoint: /increment
          header:
            X-Request-ID: abc
          query:
            dry_run: "false"
          body:
            amount: 1
        expect:
          status: 200
          body:
            count: 1

      - request:
          connection: api
          method: POST
          endpoint: /users
        expect:
          status: 201
          save:
            - field: id
              as: user_id   # available as {user_id} in subsequent steps

      - request:
          connection: api
          method: GET
          endpoint: /users/{user_id}
        expect:
          status: 200
```

> gRPC steps use the same `request:` shape — `endpoint` is the full method path (e.g. `/pkg.MyService/Method`), `connection` must resolve to a `grpc` connection, and `expect.code` is the gRPC status name.

See the [testserver example](examples/testserver/) for a working in-process server test using both the Go API and YAML loading.

---

## Matchers

Use matchers anywhere a body field value appears in Go assertions.

```go
expect.POST("/search").
    ExpectStatus(200).
    ExpectBody(map[string]any{
        "results": expect.Length(3),
        "cursor":  expect.NotEmpty{},
        "query":   expect.Contains("alice"),
        "score":   expect.Gt(0.5),
    })
```

| Matcher | Assertion |
|---------|-----------|
| `Contains(s)` | String contains substring |
| `Matches(re)` | String matches regular expression |
| `NotEmpty{}` | Value is non-nil and non-zero |
| `Gt(n)` | `actual > n` |
| `Gte(n)` | `actual >= n` |
| `Lt(n)` | `actual < n` |
| `Lte(n)` | `actual <= n` |
| `Length(n)` | Slice, array, map, or string has exactly n elements |
| `AnyOf([]int{...})` | HTTP status code is one of the given codes |

Body matching is always **partial** — expected keys must be present and match, but extra keys in the response are ignored. Array matching checks that every expected element exists somewhere in the actual array.

---

## Variables

Steps communicate through a per-scenario `VarStore`. Save a field from one response, reference it as `{key}` in any subsequent path, header, query param, or body.

```go
// Step 1: save
expect.POST("/sessions").ExpectStatus(201).Save("token", "auth_token"),

// Step 2: consume
expect.GET("/profile").WithHeader("Authorization", "Bearer {auth_token}").ExpectStatus(200),
```

Unknown `{key}` placeholders are passed through unchanged. Variables are scoped to the scenario — each scenario starts with a fresh store.

---

## Connections

```go
// HTTP — custom client or timeout
conn := &expect.HTTPConnection{
    Name:    "api",
    URL:     "https://example.com",
    Timeout: 5 * time.Second,
    Client:  myHTTPClient,
}

// gRPC — insecure by default; pass grpc.DialOption to configure TLS or interceptors
conn := expect.GRPC("svc", "localhost:50051",
    grpc.WithTransportCredentials(creds),
)

suite.WithConnections(conn)
```

Multi-connection suites route steps by connection name; the first registered connection is the default for steps that don't specify one.
