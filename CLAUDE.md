# Claude Instructions

## Interview mode
Go REST API backend. 90-minute live interview. Minimal output. Explain reasoning briefly.

**Big change** = touches more than 2 files, changes an API contract, or changes a DB schema. Ask before making one.

**Pacing**: design max 15 minutes, working code first, then tests, then polish.

**Pair mode**: one step at a time. Before writing code, say in one line what you will write and why. After, explain each non-obvious line in one line. Stop and wait for my go-ahead before the next step.

**Defaults**: standard library only (`net/http` with Go 1.22 `ServeMux` method patterns, `encoding/json`, `log/slog`). In-memory store behind an interface unless a DB is specified. Add a dependency only if the task requires it.

**Commands**: `go build ./... && go vet ./... && go test ./...`. Run `gofmt` on changed files. Go 1.22+. Add `-race` only if the code uses goroutines and a C compiler is available (on Windows, `-race` needs cgo and gcc).


## Go Backend Standards

**Error handling** (never ignore; specific types):
```go
// BAD — ignored error
result, _ := doThing()

// BAD — generic message, lost cause
result, err := doThing()
if err != nil {
    return fmt.Errorf("failed")
}

// GOOD — wrapped cause, sentinel for callers
var ErrUserNotFound = errors.New("user not found")

err := db.QueryRowContext(ctx, q, id).Scan(&user)
if errors.Is(err, sql.ErrNoRows) {
    return ErrUserNotFound
}
if err != nil {
    return fmt.Errorf("load user %q: %w", id, err)
}
```
- Wrap with `%w` when callers need `errors.Is` / `errors.As`. Prefer sentinel or typed errors over string matching.
- No panics in production paths. Map domain errors to HTTP at handlers — do not expose internals.

**Parameterized queries**: `db.Query("SELECT ... WHERE id = ?", id)`. Placeholder depends on the driver (`$1` for Postgres). Never `fmt.Sprintf` for SQL.

**Context**: Pass `context.Context` as first arg. Use for cancellation, timeouts. `context.WithTimeout`, `context.WithCancel`.

**Table-driven tests**:
```go
for _, tc := range []struct{ name string; input int; want int }{
    {"zero", 0, 0},
    {"positive", 1, 1},
} {
    t.Run(tc.name, func(t *testing.T) {
        got := fn(tc.input)
        if got != tc.want { t.Errorf("fn(%d) = %d, want %d", tc.input, got, tc.want) }
    })
}
```

**Concurrency checks**: If the code uses goroutines, run tests with `-race` when a C compiler is available. Detect goroutine leaks with `goleak`.


## Security

- Validate all input at the handler boundary: types, ranges, lengths. Reject bad input before it reaches services or the DB.
- Never hardcode or log secrets; read them from environment/config.
- Check authorization on every protected route, not just authentication.


## API Contract Standards

Apply these only when an API contract changes.

**Validation status**: Validation errors return 400. Not found returns 404.

**Error payloads**: Keep error shape consistent across endpoints (code/message/details or project equivalent). Never expose stack traces or internal exception data.

**Versioning**: Prefer explicit API versioning for breaking changes (`/v1`, `/v2` or documented equivalent). Deprecate with overlap and migration guidance.

**Compatibility**: Classify changes as compatible/risky/breaking before shipping contract updates. For breaking changes, include migration path and consumer impact.

**Documentation sync**: Update OpenAPI/spec and docs when route shape, status codes, or validation behavior changes.


## Token Efficiency

**Tool use**: Batch related reads/searches. Do not re-read files already in context unless they changed. Prefer grep for "where is X" over opening many files.

**Replies**: No filler, no repeating the user prompt. Lead with the answer or plan. Lists over long prose when enumerating steps or options.

**Code in chat**: Do not paste unchanged files. For new code, only the diff or the minimal complete block.

**Scope**: Match depth to request — a one-line fix does not need architecture essays. Defer only unrelated edge cases.

**Session hygiene**: Close obsolete topics explicitly if switching tasks to avoid carrying stale assumptions.
