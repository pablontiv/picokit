# picokit

A zero-dependency Go module providing small, focused utility packages.

Module path: `github.com/pablontiv/picokit`

## Packages

| Package | Description |
|---------|-------------|
| `autoupdate` | Parameterized staged async binary updater with opt-in version policies |
| `coverage` | Go coverage profile parser and per-package floor checker |
| `diag` | Diagnostic utilities for system and environment inspection |
| `diff` | Utilities for computing and comparing differences |
| `fuzzy` | Fuzzy matching and searching algorithms |
| `hashfile` | Utilities for computing and verifying file hashes |
| `output` | Output formatting and presentation utilities |
| `pathsec` | Path security and validation utilities |

## Features

- Zero external dependencies — each package imports only the Go standard library
- Small and focused — each package solves a specific problem
- Well-tested — 85%+ coverage on all packages
- Community-friendly — clear APIs, comprehensive docs, and maintainable code

## Getting Started

```bash
go get github.com/pablontiv/picokit
```

Then import the packages you need:

```go
import "github.com/pablontiv/picokit/pathsec"
```

## Autoupdate version policies

By default, `autoupdate` preserves its original behavior: it stages and applies
any newer release, including a new major version. Opt into compatibility-aware
automatic updates by assigning `SameMajorOnly`:

```go
updater := autoupdate.New("owner/repository", "binary")
updater.VersionPolicy = autoupdate.SameMajorOnly
```

The policy is checked before release assets are downloaded and again before an
already-staged binary is applied. A withheld update returns an
`*autoupdate.UpdateWithheldError`, whose `CurrentVersion` and `CandidateVersion`
fields let the application explain the required migration to the user.

For stable versions, `SameMajorOnly` permits newer versions with the same major
version. For `v0.x`, Semantic Versioning allows breaking changes in each minor
version, so the policy permits patch updates only within the same minor version.

## Coverage tooling

picokit ships `pkcov`, a CLI that implements [coverage-spec v1.1](docs/coverage-spec.md) — a uniform 85% floor policy applied automatically to every package in the module.

```bash
# Build the tool
go build -o pkcov ./cmd/pkcov/

# Per-package report
pkcov report --profile coverage.out --module github.com/pablontiv/picokit

# Gate (exits 1 on violations)
pkcov check --floors .coverage-floors.toml --module github.com/pablontiv/picokit

# Machine-readable output
pkcov check --output json ...
```

Minimal `.coverage-floors.toml` (v1.1 — no `packages` list needed):

```toml
default = 85
# exclude = ["cmd/experimental"]  # optional, with justification
```

Repos that adopt this policy declare compliance in their `CLAUDE.md`:

```
Coverage policy: complies with github.com/pablontiv/picokit coverage-spec v1.1
```

See [docs/coverage-spec.md](docs/coverage-spec.md) for the full contract.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and workflow.

## License

picokit is licensed under the [Apache License 2.0](LICENSE). See [LICENSE](LICENSE) for terms.

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.
