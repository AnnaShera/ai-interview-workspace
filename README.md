# ai-interview-workspace

Claude Code setup for a 90-minute Go interview. Minimal output, ask before big changes.

## Active agents (`.claude/agents/`)

Use in this order:

1. **architect** (design): pick structure and tradeoffs before coding. Read-only.
2. **backend-go** (code): implement handlers, services, data access.
3. **tester** (tests): table-driven tests with `httptest`, run `go test -cover -race ./...`.
4. **reviewer** (final check): review the diff before calling it done. Read-only.

## Reserve (`reserve/`)

Move into `.claude/agents/` only when needed. Agents load at startup, so decide on reserve agents before starting Claude Code.

- **security**: the task handles user input, auth, or secrets and needs a vulnerability pass.
- **pm**: the requirement is vague and needs splitting into tasks with acceptance criteria.
- **sql-database**: the task needs a relational schema, migration, or slow-query tuning.
- **nosql-database**: the task needs keys, partitioning, or access-pattern modeling (Mongo, Dynamo, Redis).
- **performance**: an endpoint or algorithm is slow or has a latency target (pprof, benchmarks).

```
mv reserve/security.md .claude/agents/
```

## Copy into another project

```
cp -r .claude CLAUDE.md /path/to/project/
```

PowerShell:

```
Copy-Item -Recurse .claude, CLAUDE.md C:\path\to\project\
```

This overwrites the target's `CLAUDE.md` if it has one.

Copy one reserve agent into another project:

```
cp reserve/security.md /path/to/project/.claude/agents/
```

PowerShell:

```
Copy-Item reserve\security.md C:\path\to\project\.claude\agents\
```
