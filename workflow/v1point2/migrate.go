package v1point2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// WorkflowFromReaders gets a workflow from supplied readers.
func MigrateFromReaders(rs ...io.ReadCloser) (*Workflow, error) {
	for _, r := range rs {
		defer r.Close()
		wf := &Workflow{}
		dec := json.NewDecoder(r)
		if err := dec.Decode(wf); err != nil {
			return nil, err
		}
		fmt.Printf("Read workflow %s \n", wf.Properties.Description)
		fmt.Printf("Read workflow version %s \n", wf.Version)
		return wf, nil
	}
	return nil, errors.New("no files were supplied")
}
