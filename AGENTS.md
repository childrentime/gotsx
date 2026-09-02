# Working on the gotsx repository

This is the framework itself (compiler, runtime, client runtime, CLI, design system, demos), not an app built
with it. The working notes, language policy, commands and invariants live in **`CLAUDE.md`** — read it first.

Apps *built* with gotsx get their own `AGENTS.md` from `gotsx new` (a managed block that points agents at the
version-matched docs in `app/.gen/docs/`); the source of that block and of those docs is `cmd/gotsx/agent.go`
and `cmd/gotsx/docs/*.md`. When you change the language, a convention or an error message, update those docs
in the same change.

Quick loop: `make test-fast` (unit) → `make gen && make check` (every demo compiles) → `make test` (everything,
including the scaffold end-to-end) → browser-verify demos with Playwright when the client runtime or a demo changed.
