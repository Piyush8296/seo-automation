# CLAUDE.md

## Project

Go-based SEO automation tool with a web UI. CLI commands in `cmd/`, business logic in `internal/`, frontend in `ui/`.

## Workflow Rules

- **One fix/feature per conversation.** Do not mix unrelated changes in a single session.
- **Read only what's relevant.** Before starting, identify the minimal set of files needed. Do not explore the entire codebase.
- **Keep commits atomic.** Each commit should address exactly one concern — a single bug fix, a single feature, or a single refactor.
- **Do not refactor surrounding code** unless it is directly required by the task at hand.
- **Avoid loading large files** (reports, generated output, vendored code) into context unless explicitly needed.

## Build & Run

```bash
go build -o seo-audit .          # build
go run main.go audit              # run audit
go run main.go server             # start web server
make build                        # via Makefile
```

## Test

```bash
go test ./...                     # all tests
go test ./internal/checks/...     # specific package
```
