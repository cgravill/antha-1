FROM eu.gcr.io/antha-images/golang:1.12.4-build AS build
ARG NETRC
RUN printf "%s\n" "$NETRC" > /root/.netrc
RUN set -ex && wget -O - -q https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.16.0
RUN mkdir -p /go/src/github.com/antha-lang/antha
WORKDIR /go/src/github.com/antha-lang/antha
ADD . /go/src/github.com/antha-lang/antha
RUN set -ex && go get -t ./...
RUN set -ex && make lint
