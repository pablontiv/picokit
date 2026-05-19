# picokit

A zero-dependency Go module providing small, focused utility packages.

Module path: `github.com/pablontiv/picokit`

## Packages

| Package | Description |
|---------|-------------|
| `autoupdate` | Parameterized staged async binary updater |
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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and workflow.

## License

picokit is licensed under the [PolyForm Noncommercial License 1.0.0](LICENSE). See [LICENSE](LICENSE) for terms.

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.
