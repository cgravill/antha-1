FROM eu.gcr.io/antha-images/golang:1.12.5-build AS build
ARG COMMIT_SHA
ARG NETRC
RUN printf "%s\n" "$NETRC" > /root/.netrc
RUN mkdir /antha
WORKDIR /antha
RUN set -ex && go mod init antha && go mod edit "-require=github.com/Synthace/antha@$COMMIT_SHA" && go mod download
RUN set -ex && go build github.com/Synthace/antha/...
RUN set -ex && go install github.com/Synthace/antha/cmd/...
RUN set -ex && go test -c github.com/Synthace/antha/cmd/elements
COPY scripts/*.sh /antha/

FROM eu.gcr.io/antha-images/golang:1.12.5-build AS lint
COPY --from=build /root/.netrc /root/.cache /root/
COPY --from=build /go /go
COPY --from=build /antha /antha
WORKDIR /antha
RUN set -ex && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.16.0
RUN set -ex && cp -a $(go list -f '{{ .Dir }}' github.com/Synthace/antha) /lintme
WORKDIR /lintme
RUN set -ex && go mod edit "-dropreplace=github.com/Synthace/antha-runner" "-dropreplace=github.com/Synthace/instruction-plugins" && go mod download && golangci-lint run --deadline=5m -E gosec -E gofmt

FROM eu.gcr.io/antha-images/golang:1.12.5-build AS tests
ARG COMMIT_SHA
ARG BRANCH_NAME
ARG COVERALLS_TOKEN
ARG BUILD_ID
COPY --from=build /root/.netrc /root/.cache /root/
COPY --from=build /go /go
COPY --from=build /antha /antha
WORKDIR /antha
RUN ./antha-test.sh "$COVERALLS_TOKEN" "$COMMIT_SHA" "$BRANCH_NAME" "$BUILD_ID"

FROM eu.gcr.io/antha-images/golang:1.12.5-build AS cloud
## This target produces an image that is used both for gitlab elements CI, and also workflow execution in the cloud
COPY --from=tests /root/.cache /root/
COPY --from=tests /go /go
COPY --from=build /antha /antha
WORKDIR /antha
# Do these builds to pre-warm the build cache. This makes a HUGE difference to performance in the cloud
RUN set -ex && go build github.com/Synthace/antha/... github.com/Synthace/antha-runner/... github.com/Synthace/instruction-plugins/...
# These are for the gitlab CI for elements:
ONBUILD ADD . /elements
