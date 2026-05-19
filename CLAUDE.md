# picokit

Go module providing small, zero-dependency utility packages for use across pablontiv projects.

Module path: `github.com/pablontiv/picokit`

## Packages

_(populated by T002–T005)_

## Conventions

- Zero external dependencies per package — each package in `picokit/` must compile with only stdlib.
- Tests live alongside the code (`_test.go` files in the same package).
- Coverage threshold: 85%.
- CI delegates to `pablontiv/crossbeam@v1` (go-ci, gitleaks, go-release, codeql, scorecard).

## Local commands

```
just check    # gofmt + go vet
just test     # go test ./...
just fmt      # gofmt -w
just coverage-summary  # go test -cover ./...
just audit    # go mod verify
```
