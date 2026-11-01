# Claude Instructions

## Interview mode
Interview mode. 90 minutes. Minimal output. Ask before big changes. Explain reasoning briefly.


# Go Backend Standards

**Error handling** (never ignore; specific types):
```go
// ❌ BAD — ignored error, generic message, lost cause
result, _ := doThing()
if err != nil {
    return fmt.Errorf("failed")
}

// ✅ GOOD — wrapped cause, sentinel for callers
var ErrUserNotFound = errors.New("user not found")

row, err := db.QueryRowContext(ctx, q, id).Scan(&user)
if errors.Is(err, sql.ErrNoRows) {
    return ErrUserNotFound
}
if err != nil {
    return fmt.Errorf("load user %q: %w", id, err)
}
```
- Wrap with `%w` when callers need `errors.Is` / `errors.As`. Prefer sentinel or typed errors over string matching.
- No panics in production paths. Map domain errors to HTTP at handlers — do not expose internals.

**Parameterized queries**: `db.Query("SELECT ... WHERE id = ?", id)`. Never `fmt.Sprintf` for SQL. Use `sqlx.Named` for named params.

**Context**: Pass `context.Context` as first arg. Use for cancellation, timeouts. `context.WithTimeout`, `context.WithCancel`.

**Table-driven tests**:
```go
for _, tc := range []struct{ name string; input int; want int }{
    {"zero", 0, 0},
    {"positive", 1, 1},
} {
    t.Run(tc.name, func(t *testing.T) {
        got := fn(tc.input)
        if got != tc.want { t.Errorf("got %d", got) }
    })
}
```

**Profiling**: `go test -bench=. -cpuprofile=cpu.prof`. `pprof -http=:8080 cpu.prof`. Check goroutine leaks with `-trace`.


# API Contract Standards

**Validation status**: Validation error status must follow the project API contract. If unspecified, use framework default and document it explicitly.

**Error payloads**: Keep error shape consistent across endpoints (code/message/details or project equivalent). Never expose stack traces or internal exception data.

**Versioning**: Prefer explicit API versioning for breaking changes (`/v1`, `/v2` or documented equivalent). Deprecate with overlap and migration guidance.

**Compatibility**: Classify changes as compatible/risky/breaking before shipping contract updates. For breaking changes, include migration path and consumer impact.

**Documentation sync**: Update OpenAPI/spec and docs when route shape, status codes, or validation behavior changes.


# Token Efficiency

**Inputs**: Ask one focused question per turn when possible. Use `@file` / `@folder` / symbols instead of pasting large code. Point to paths and line ranges instead of quoting whole files.

**Tool use**: Batch related reads/searches. Do not re-read files already in context unless they changed. Prefer grep/codebase_search for "where is X" over opening many files. For MCP vs shell, follow `mcp-integrations.mdc` CLI routing.

**Replies**: No filler, no repeating the user prompt. Lead with the answer or plan. Lists over long prose when enumerating steps or options.

**Code in chat**: Do not paste unchanged files. For existing code, use editor line-range citations. For new code, only the diff or the minimal complete block.

**Scope**: Match depth to request — a one-line fix does not need architecture essays. Defer edge cases unless asked.

**Session hygiene**: Close obsolete topics explicitly if switching tasks to avoid carrying stale assumptions.
