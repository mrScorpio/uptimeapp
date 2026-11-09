# Backend Contributor Guide

## Structure

The Go module is defined in `backend/go.mod`. Keep executable entry points in `cmd/`, private application code in `internal/`, reusable public packages in `pkg/`, configuration templates in `configs/`, and operational documentation in `docs/`. Place tests beside the package they cover, such as `internal/health/checker_test.go`.

## Commands

Run these commands from `backend/`:

- `find . -name '*.go' -print0 | xargs -0 gofmt -w` formats Go files.
- `go test ./...` runs all tests.
- `go build ./...` checks that all packages build.

## Style & Naming

Use `gofmt` and standard Go naming. Exported identifiers use PascalCase; internal identifiers use camelCase; package names are short, lowercase words. Keep packages focused and prefer explicit error handling.

## Testing

Use Go’s standard `testing` package. Name files `_test.go` and functions `TestThing` or `TestThing_WhenCondition`. Add focused tests for new behavior and run `go test ./...` before review.
