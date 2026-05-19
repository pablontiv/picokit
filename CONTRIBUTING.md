# Contributing to picokit

Thank you for your interest in contributing!

picokit is a zero-dependency Go library providing small utility packages for use across projects.

## Development Setup

```bash
# Clone and enter the repo
git clone https://github.com/pablontiv/picokit.git
cd picokit

# Set up git hooks
git config core.hooksPath .githooks

# Verify environment
go build ./...
go test ./...
golangci-lint run ./...
```

Requires Go 1.21+ and [golangci-lint v2](https://golangci-lint.run/).

## Workflow

1. Fork the repository
2. Create a feature branch from `master`
3. Make your changes
4. Run `go test ./...` and `golangci-lint run ./...`
5. Commit using [Conventional Commits](https://www.conventionalcommits.org/)
6. Open a Pull Request

## Releasing

Releases are fully automated via CI. On push to `master`, CI analyzes conventional commit prefixes, calculates the next semver version, and creates a GitHub Release. No manual release steps needed.

## Commit Convention

```
type(scope): description
```

| Type | When to use |
|------|-------------|
| `feat` | New package or public API function |
| `fix` | Bug fix |
| `refactor` | Internal restructuring, no behavior change |
| `test` | Adding or updating tests |
| `docs` | Documentation only |
| `chore` | Build, CI, dependency updates |

Breaking changes use `!` suffix: `feat!: change package API`

## Code Style

- **Formatting**: `gofmt` (enforced in CI and pre-commit hook)
- **Linting**: `golangci-lint v2` (CI lint gate)
- **Testing**: stdlib `testing` package; unit tests required
- **Dependencies**: zero external dependencies per package — all packages must compile with only stdlib

## Package Guidelines

- Each package in `picokit/` must be independently useful
- Coverage threshold: 85%
- Document exported functions and types with comments
- Keep packages focused and small

## Reporting Issues

- **Bugs**: Use the bug report template
- **Features**: Use the feature request template
- **Security**: See [SECURITY.md](SECURITY.md) for responsible disclosure
