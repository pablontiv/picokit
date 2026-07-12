# Zero-dependency compliance and CI-delegation debt — Design

Date: 2026-07-12
Status: Approved (autonomous session, user-delegated)

## Problem

1. The `coverage` package imports `github.com/pelletier/go-toml/v2` (`coverage/floors.go`),
   violating the project convention "Zero external dependencies per package — each package
   must compile with only stdlib" (CLAUDE.md) and the README claim "zero-dependency Go module".
2. `cmd/pkcov` imports `github.com/spf13/cobra` for three trivial subcommands, violating the
   same convention.
3. The `coverage-gate` job in `.github/workflows/ci.yml` uses `actions/checkout` and
   `actions/setup-go` directly instead of delegating to `pablontiv/crossbeam@v1` like every
   other job — undocumented architectural drift.

## Goals

- `go.mod` ends with zero `require` entries (module compiles with stdlib only).
- `.coverage-floors.toml` file format is preserved exactly — it is a public contract consumed
  by CI in this and future repos.
- `pkcov` CLI surface is preserved exactly: `pkcov check --profile <path> --floors <path>`,
  `pkcov report --profile <path>`, `pkcov version`. CI invokes `check`; changing flags breaks CI.
- The coverage-gate exception is documented where the next reader will look.

## Non-goals

- Full TOML 1.0 compliance. The parser supports only the grammar `.coverage-floors.toml` uses.
- Moving the coverage gate into crossbeam (`pkcov-gate.yml` reusable workflow). Deferred until
  a second repo adopts pkcov; noted as future roadmap item.

## Design

### 1. Minimal TOML-subset parser (`coverage/toml.go`)

New unexported function `parseFloorsTOML(data []byte) (*Floors, error)` replacing
`toml.Unmarshal`. Supported grammar (exactly what the config contract needs):

- Comments: `#` to end of line.
- Blank lines.
- `key = <integer>` — for `default`.
- `key = [ "s1", "s2" ]` — single-line string arrays for `packages` / `exclude`;
  basic double-quoted strings, no escape sequences beyond `\"` and `\\`.
- Unknown keys: ignored (matches go-toml's lenient behavior for this struct).
- Anything else (tables, multi-line arrays, other value types): explicit error naming the
  line number, so a config drifting outside the subset fails loudly instead of silently.

`LoadFloors` keeps its exact signature and validation (`default > 0`).

Error contract: `parse floors: line N: <reason>` wrapped in the existing
`parse floors: %w` from `LoadFloors`.

### 2. cobra → stdlib `flag` (`cmd/pkcov`)

Hand-rolled subcommand dispatch in `main.go`/`root.go`: read `os.Args[1]`, switch over
`check` / `report` / `version`, each with its own `flag.FlagSet` (ContinueOnError) carrying the
same flag names and defaults as today. Unknown subcommand or `-h`/`--help` prints usage to
stderr and exits 2 (flag package convention). Exit codes for `check` violations unchanged.

### 3. Documentation

- `ci.yml`: comment above `coverage-gate` explaining it intentionally does not delegate to
  crossbeam (dogfoods this repo's own pkcov; a reusable workflow would be premature).
- `CLAUDE.md`: one line in Conventions noting the coverage-gate exception.
- `go.mod`/`go.sum`: drop go-toml, cobra, and transitive deps (`go mod tidy`).

## Testing

Strict TDD. New table-driven tests for the parser (`coverage/toml_test.go`): valid config,
comments, arrays, unknown keys, and error cases (bad int, unterminated string, table header,
multi-line array). Existing `LoadFloors` tests and `cmd/pkcov/pkcov_test.go` must pass
unchanged — they are the regression harness for the contract. Coverage floor: 85 per package.

## Risks

- Parser under-covers a TOML feature someone later uses in a floors file → mitigated by
  loud line-numbered errors and the documented subset.
- CLI flag drift breaks CI → mitigated by existing pkcov tests plus CI itself running
  `pkcov check` on this PR.
