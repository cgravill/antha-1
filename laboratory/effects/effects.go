package effects

import (
	"github.com/Synthace/antha/instructions"
	"github.com/Synthace/antha/inventory"
	"github.com/Synthace/antha/inventory/cache/plateCache"
	"github.com/Synthace/antha/laboratory/effects/id"
	"github.com/Synthace/antha/microArch/sampletracker"
	"github.com/Synthace/antha/workflow"
)

type LaboratoryEffects struct {
	FileManager   *FileManager
	Trace         *instructions.Trace
	Maker         *instructions.Maker
	SampleTracker *sampletracker.SampleTracker
	Inventory     *inventory.Inventory
	// TODO the plate cache should go away and the use sites should be rewritten around plate types.
	PlateCache  *plateCache.PlateCache
	IDGenerator *id.IDGenerator
}

func NewLaboratoryEffects(wf *workflow.Workflow, simId workflow.BasicId, fm *FileManager) *LaboratoryEffects {
	idGen := id.NewIDGenerator(string(simId))
	le := &LaboratoryEffects{
		FileManager:   fm,
		Trace:         instructions.NewTrace(),
		Maker:         instructions.NewMaker(),
		SampleTracker: sampletracker.NewSampleTracker(),
		Inventory:     inventory.NewInventory(idGen),
		IDGenerator:   idGen,
	}
	le.PlateCache = plateCache.NewPlateCache(le.Inventory.Plates)
	le.Inventory.LoadForWorkflow(wf)

	return le
}
