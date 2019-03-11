package v1point2

import (
	"encoding/json"
	"io"
	"os"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
)

type Migrater struct {
	Logger       *logger.Logger
	BaseWorkflow *workflow.Workflow
	OldWorkflow  *Workflow
}

// NewMigrater creates a new migration object
func NewMigrater(logger *logger.Logger, merges []string, migrate io.ReadCloser) (*Migrater, error) {

	if wf, err := baseWorkflow(merges); err != nil {
		return nil, err
	} else if owf, err := readWorkflow(migrate); err != nil {
		return nil, err
	} else {
		return &Migrater{
			Logger:       logger,
			BaseWorkflow: wf,
			OldWorkflow:  owf,
		}, nil
	}
}

// MigrateAll perform all migration steps
func (m *Migrater) MigrateAll() error {
	if err := m.migrateParameters(); err != nil {
		return err
	} else if err := m.migrateElements(); err != nil {
		return err
	}

	return nil
}

func (m *Migrater) migrateElements() error {
	for k := range m.OldWorkflow.Processes {
		name := workflow.ElementInstanceName(k)
		if ei, err := m.OldWorkflow.MigratedElement(k); err != nil {
			return err
		} else {
			m.BaseWorkflow.Elements.Instances[name] = ei
		}
	}

	return nil
}

func (m *Migrater) migrateParameters() error {
	m.BaseWorkflow.JobId = workflow.JobId(m.OldWorkflow.Properties.Name)

	if m.BaseWorkflow.Meta.Rest == nil {
		m.BaseWorkflow.Meta.Rest = make(map[string]interface{})
	}

	m.BaseWorkflow.Meta.Rest["Description"] = m.OldWorkflow.Properties.Description

	return nil
}

func (m *Migrater) WriteTo(target string) error {
	if target == "" {
		_, err := m.BaseWorkflow.WriteTo(os.Stdout)
		return err
	} else {
		return m.BaseWorkflow.WriteToFile(target)
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
