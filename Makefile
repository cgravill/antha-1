all: build lint test testelements

build:
	go build ./...
	go install ./cmd/...

lint:
	golangci-lint run --deadline=5m -E gosec -E gofmt

test:
	./scripts/antha-test.sh

testelements: cmd/composer/repositories.json
	go test github.com/antha-lang/antha/cmd/elements -v -args -keep $(abspath $<)

.PHONY: all build lint test testelements
