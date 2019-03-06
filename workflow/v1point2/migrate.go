package v1point2

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
)

// MigrateWorkflow moves an existing
func MigrateWorkflow(logger *logger.Logger, supportWorkflows []string, migrate string, migrateTo string) (string, error) {
	migrateTo = correctMigrateTo(migrate, migrateTo)
	logger.Log(fmt.Sprintf("Migrating %s, output to %s\n", migrate, migrateTo))

	if baseWf, err := baseWorkflow(supportWorkflows); err != nil {
		return "", err
	} else if _, err := readWorkflow(migrate); err != nil {
		return "", err
	} else {
		baseWf.WriteToFile(migrateTo)
		return "", errors.New("failed to migrate file")
	}
}

func baseWorkflow(paths []string) (*workflow.Workflow, error) {
	if len(paths) == 0 {
		return &workflow.Workflow{}, nil
	}

	if rs, err := workflow.ReadersFromPaths(paths); err != nil {
		return nil, err
	} else if wf, err := workflow.WorkflowFromReaders(rs...); err != nil {
		return nil, err
	} else {
		return wf, nil
	}
}

func correctMigrateTo(migrate string, migrateTo string) string {
	if migrateTo == "" {
		return filepath.Join(filepath.Dir(migrate), "migrated_"+filepath.Base(migrate))
	} else {
		return migrateTo
	}
}

func readWorkflow(path string) (*Workflow, error) {
	if r, err := os.Open(path); err != nil {
		return nil, err
	} else {
		defer r.Close()
		wf := &Workflow{}
		dec := json.NewDecoder(r)
		if err := dec.Decode(wf); err != nil {
			return nil, err
		} else {
			return wf, nil
		}
	}
}
