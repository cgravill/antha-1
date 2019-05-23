all: build lint test

build:
	go build ./...
	go install ./cmd/...

lint:
	golangci-lint run --deadline=5m -E gosec -E gofmt

test:
	./scripts/antha-test.sh

.PHONY: all build lint test
