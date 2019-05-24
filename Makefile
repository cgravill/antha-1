# By default, REPOSITORIES is set to cmd/composer/repositories.json,
# but you can override this on the command line, eg make testelements REPOSITORIES=path/to/my/repositories.json
# though note that the path will be interpreted as relative to the location of this Makefile
REPOSITORIES ?= cmd/composer/repositories.json

all: build lint test testelements

build:
	go build ./...
	go install ./cmd/...

lint:
	golangci-lint run --deadline=5m -E gosec -E gofmt

test:
	./scripts/antha-test.sh

testelements: $(REPOSITORIES)
	go test github.com/antha-lang/antha/cmd/elements -v -args -keep $(abspath $<)

.PHONY: all build lint test testelements
