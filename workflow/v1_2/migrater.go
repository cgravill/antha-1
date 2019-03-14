package v1_2

import (
	"encoding/json"
	"io"
	"os"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/utils"
	"github.com/antha-lang/antha/workflow"
)

type Migrater struct {
	Logger      *logger.Logger
	Cur         *workflow.Workflow
	OldWorkflow *workflowv1_2
}

// NewMigrater creates a new migration object
func NewMigrater(logger *logger.Logger, merges []string, migrate io.ReadCloser) (*Migrater, error) {

	if wf, err := baseWorkflow(merges); err != nil {
		return nil, err
	} else if owf, err := readWorkflow(migrate); err != nil {
		return nil, err
	} else {
		return &Migrater{
			Logger:      logger,
			Cur:         wf,
			OldWorkflow: owf,
		}, nil
	}
}

// MigrateAll perform all migration steps
func (m *Migrater) MigrateAll() error {
	return utils.ErrorSlice{
		m.migrateParameters(),
		m.migrateElements(),
	}.Pack()

}

func (m *Migrater) migrateElements() error {
	for k := range m.OldWorkflow.Processes {
		name := workflow.ElementInstanceName(k)
		ei, err := m.OldWorkflow.MigrateElement(k)
		if err != nil {
			return err
		}
		m.Cur.Elements.Instances[name] = ei
	}

	return nil
}

func (m *Migrater) migrateParameters() error {
	m.Cur.JobId = workflow.JobId(m.OldWorkflow.Properties.Name)
	m.Cur.Meta.InitEmpty()
	m.Cur.Meta.Rest["Description"] = m.OldWorkflow.Properties.Description
	return nil
}

func (m *Migrater) WriteToPath(target string) error {
	if target == "" || target == "-" {
		return m.Cur.WriteToStream(os.Stdout)
	}
	return m.Cur.WriteToFile(target)
}

func baseWorkflow(paths []string) (*workflow.Workflow, error) {
	if len(paths) == 0 {
		return &workflow.Workflow{}, nil
	}

	rs, err := workflow.ReadersFromPaths(paths)

	if err != nil {
		return nil, err
	}

	return workflow.WorkflowFromReaders(rs...)
}

func readWorkflow(r io.ReadCloser) (*workflowv1_2, error) {
	defer r.Close()
	wf := &workflowv1_2{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(wf); err != nil {
		return nil, err
	}
	return wf, nil
}
