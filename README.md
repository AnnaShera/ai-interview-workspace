# ai-interview-workspace

Claude Code setup for a 90-minute Go interview. Minimal output, ask before big changes.

## Active agents (`.claude/agents/`)

Use in this order:

1. **architect** (design): pick structure and tradeoffs before coding. Read-only.
2. **tester** (tests): table-driven tests with `httptest`, run `go test -cover ./...`.
3. **reviewer** (final check): review the changes. It can run git diff, go vet, and go test, but cannot edit files.

## Reserve (`reserve/`)

Move into `.claude/agents/` only when needed. Agents load at startup, so decide on reserve agents before starting Claude Code.

- **security**: the task handles user input, auth, or secrets and needs a vulnerability pass.
- **pm**: the requirement is vague and needs splitting into tasks with acceptance criteria.
- **sql-database**: the task needs a relational schema, migration, or slow-query tuning.
- **nosql-database**: the task needs keys, partitioning, or access-pattern modeling (Mongo, Dynamo, Redis).
- **performance**: an endpoint or algorithm is slow or has a latency target (pprof, benchmarks).
- **backend-go**: you want a whole handler or service written in one go, without step-by-step pair coding.

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

## Interview day checklist

1. Check `go version` is 1.22 or newer.
2. Read the task and decide on reserve agents.
3. Copy `.claude`, `CLAUDE.md`, and any reserve agent into the project.
4. Start Claude Code, then use architect for design, code in the main session, tester for tests, reviewer for the final check.

## Time plan

- 0-5: read the task, ask clarifying questions, state assumptions.
- 5-15: design with architect.
- 15-65: code in the main session, tests with tester.
- 65-80: reviewer, fix critical issues.
- 80-90: build, vet, test, then summarize tradeoffs and next steps.
