# List available recipes
default:
    @just --list

# Build the binary
build:
    go build -o grpc_client main.go

# Run all tests
test:
    go test ./...

# Run linters
lint:
    golangci-lint run

# Run linters and fix issues
lint-fix:
    golangci-lint run --fix
