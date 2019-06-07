// +build !linkedDrivers

package mixer

import "github.com/Synthace/antha/workflow"

func (bm *BaseMixer) maybeLinkedDriver(wf *workflow.Workflow, data []byte) error {
	return nil
}
