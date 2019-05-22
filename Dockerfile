FROM eu.gcr.io/antha-images/golang:1.12.4-build AS build
ARG COMMIT_SHA
ARG NETRC
RUN printf "%s\n" "$NETRC" > /root/.netrc
RUN mkdir /antha
WORKDIR /antha
RUN set -ex && go mod init antha && go mod edit "-require=github.com/antha-lang/antha@$COMMIT_SHA" && go mod download
RUN set -ex && go install github.com/antha-lang/antha/cmd/...
RUN set -ex && go test -c github.com/antha-lang/antha/cmd/elements
COPY scripts/*.sh /antha/

FROM eu.gcr.io/antha-images/golang:1.12.4-build AS lint
COPY --from=build /root/.netrc /root/.cache /root/
COPY --from=build /go /go
COPY --from=build /antha /antha
WORKDIR /antha
RUN set -ex && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.16.0
RUN set -ex && cp -a $(go list -f '{{ .Dir }}' github.com/antha-lang/antha) /lintme
WORKDIR /lintme
RUN set -ex && go mod edit "-dropreplace=github.com/Synthace/antha-runner" "-dropreplace=github.com/Synthace/instruction-plugins" && golangci-lint run --deadline=5m -E gosec -E gofmt

FROM eu.gcr.io/antha-images/golang:1.12.4-build AS tests
ARG COVERALLS_TOKEN
COPY --from=build /root/.netrc /root/.cache /root/
COPY --from=build /go /go
COPY --from=build /antha /antha
WORKDIR /antha
RUN ./antha-test.sh "$COVERALLS_TOKEN"

FROM eu.gcr.io/antha-images/golang:1.12.4-build AS cloud
## This target produces an image that is used both for gitlab elements CI, and also workflow execution in the cloud
COPY --from=tests /root/.cache /root/
COPY --from=tests /go /go
COPY --from=build /antha /antha
WORKDIR /antha
# Do these builds to pre-warm the build cache. This makes a HUGE difference to performance in the cloud
RUN set -ex && go build github.com/antha-lang/antha/... github.com/Synthace/antha-runner/... github.com/Synthace/instruction-plugins/...
# These are for the gitlab CI for elements:
ONBUILD ADD . /elements
