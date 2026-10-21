.PHONY: build test check
build:
	go build -trimpath -o bin/boosty-mcp ./cmd/boosty-mcp
test:
	go test -race ./...
check: test
	go vet ./...
	test -z "$$(gofmt -l cmd internal)"
