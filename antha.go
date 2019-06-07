//go:generate protoc -I${GOPATH}/src ${GOPATH}/src/github.com/Synthace/antha/protobuf/protobuf.proto --go_out=${GOPATH}/src
//go:generate clang-format -i protobuf/protobuf.proto

package antha
