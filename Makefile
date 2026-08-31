.PHONY: run test build tidy

run:
	go run ./cmd/mcp

test:
	go test ./...

build:
	go build -o bin/gas-up-mcp ./cmd/mcp

tidy:
	go mod tidy
