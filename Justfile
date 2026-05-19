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

# Coverage summary
coverage-summary:
    go test -cover ./...

# Audit dependencies
audit:
    go mod verify
