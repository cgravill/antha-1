package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Synthace/antha-runner/protobuf"
	"github.com/golang/protobuf/proto"
)

// A simple command-line tool for converting a SimulateRequest protobuf file to
// JSON, so that humans can read it.
//
// Usage: `dumpjson request.pb`
func main() {
	cmd := filepath.Base(os.Args[0])
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v PROTOBUF_FILE", cmd)
	}
	pbFilePath := os.Args[1]

	p := &protobuf.SimulateRequest{}
	bs, err := ioutil.ReadFile(pbFilePath)
	if err != nil {
		panic(err)
	}
	err = proto.Unmarshal(bs, p)
	if err != nil {
		panic(err)
	}

	bs, err = json.Marshal(p)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", bs)
}
