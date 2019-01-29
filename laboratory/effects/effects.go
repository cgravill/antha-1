package effects

import (
	"github.com/antha-lang/antha/inventory"
	"github.com/antha-lang/antha/inventory/cache/plateCache"
	"github.com/antha-lang/antha/laboratory/effects/id"
	"github.com/antha-lang/antha/microArch/sampletracker"
)

type LaboratoryEffects struct {
	JobId string

	Trace         *Trace
	Maker         *Maker
	SampleTracker *sampletracker.SampleTracker
	Inventory     *inventory.Inventory
	// TODO the plate cache should go away and the use sites should be rewritten around plate types.
	PlateCache  *plateCache.PlateCache
	IDGenerator *id.IDGenerator
}

func NewLaboratoryEffects(jobId string) *LaboratoryEffects {
	idGen := id.NewIDGenerator(jobId)
	le := &LaboratoryEffects{
		JobId:         jobId,
		Trace:         NewTrace(),
		Maker:         NewMaker(),
		SampleTracker: sampletracker.NewSampleTracker(),
		Inventory:     inventory.NewInventory(idGen),
		IDGenerator:   idGen,
	}
	le.PlateCache = plateCache.NewPlateCache(le.Inventory.PlateTypes)

	return le
}
