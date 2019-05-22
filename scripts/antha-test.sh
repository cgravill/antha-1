#! /bin/bash

## This script is run by cloudbuild as part of the CI for antha itself.
set -o nounset -o errexit -o pipefail -o noclobber
shopt -s failglob

COVERALLS_TOKEN=${1:-}
COMMIT_SHA=${2:-}
BRANCH_NAME=${3:-}

## There are some packages that only contain test files. Go test gets
## upset if you try to include these packages in coverage, so we have
## to filter them out:
COVERPKG=$(go list -f '{{if (len .GoFiles) gt 0}}{{.ImportPath}}{{end}}' github.com/antha-lang/antha/... | tr '\n' ',' | sed -e 's/,$//')

go test -covermode=atomic -coverprofile=cover.profile -coverpkg="${COVERPKG}" github.com/antha-lang/antha/...

if [[ -n "${COVERALLS_TOKEN}" && -s cover.profile ]]; then
    coveralls -reponame="github.com/antha-lang/antha" -repotoken="${COVERALLS_TOKEN}" -commitsha="${COMMIT_SHA}" -branchname="${BRANCH_NAME}" cover.profile
fi
