package v1point2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
)

// MigrateWorkflow moves an existing
func MigrateWorkflow(logger *logger.Logger, merges []string, migrate io.ReadCloser, outfile string) (string, error) {
	logger.Log(fmt.Sprintf("Migrating %s, output to %s\n", migrate, outfile))
	if wf, err := baseWorkflow(merges); err != nil {
		return "", err
	} else if _, err := readWorkflow(migrate); err != nil {
		return "", err
	} else {
		writeTo(wf, outfile)
		return "", errors.New("failed to migrate file")
	}
}

func writeTo(wf *workflow.Workflow, target string) error {
	if target == "" {
		_, err := wf.WriteTo(os.Stdout)
		return err
	} else {
		return wf.WriteToFile(target)
	}
}

func baseWorkflow(paths []string) (*workflow.Workflow, error) {
	if len(paths) == 0 {
		return &workflow.Workflow{}, nil
	}

	if rs, err := workflow.ReadersFromPaths(paths); err != nil {
		return nil, err
	} else if wf, err := workflow.PartialWorkflowFromReaders(rs...); err != nil {
		return nil, err
	} else {
		return wf, nil
	}
}

func readWorkflow(r io.ReadCloser) (*Workflow, error) {
	defer r.Close()
	wf := &Workflow{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(wf); err != nil {
		return nil, err
	} else {
		return wf, nil
	}
}
