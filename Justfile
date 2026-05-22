set shell := ["bash", "-c"]

# Default recipe
default: check test

# Run fmt + vet
check:
    gofmt -l . | grep -q . && { echo "gofmt: unformatted files"; exit 1; } || true
    [ -z "$(go list ./... 2>/dev/null)" ] || go vet ./...

# Format code
fmt:
    gofmt -w .

# Run tests
test:
    go test ./...

# Coverage report (per-package table)
coverage:
    go test ./... -coverprofile=coverage.out -count=1
    go run ./cmd/pkcov report --profile coverage.out --module github.com/pablontiv/picokit

# Coverage gate — exits 1 if any package is below 85%
coverage-check:
    go test ./... -coverprofile=coverage.out -count=1
    go run ./cmd/pkcov check --profile coverage.out --floors .coverage-floors.toml --module github.com/pablontiv/picokit

# Audit dependencies
audit:
    go mod verify
