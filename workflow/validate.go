package workflow

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/antha-lang/antha/utils"
)

func (wf *Workflow) Validate() error {
	return utils.ErrorSlice{
		wf.SchemaVersion.Validate(),
		wf.WorkflowId.Validate(false),
		wf.Repositories.validate(),
		wf.Elements.validate(wf),
		wf.Inventory.validate(),
		wf.Config.validate(),
		wf.Simulation.validate(wf),
	}.Pack()
}

func (sv SchemaVersion) Validate() error {
	switch sv {
	case CurrentSchemaVersion:
		return nil
	default:
		return fmt.Errorf("Validation error: Invalid Schema Version. Migration required. Require version '%v'; Received version '%v'", CurrentSchemaVersion, sv)
	}
}

func (basicId BasicId) Validate(permitEmpty bool) error {
	if basicId == "" && !permitEmpty {
		return errors.New("Invalid Id: may not be empty")
	}
	// We rely on the json schema to enforce further value restrictions
	// (i.e. there's a pattern in there - see
	// workflow/schemas/workflow.schema.json)
	return nil
}

func (rs Repositories) validate() error {
	// We have to enforce that all repositories are not only unique, but that no
	// one repository is a prefix of another. To enforce this, we sort the
	// prefixes (so shortest will come first) and then we need to only test
	// against the tail of the list.
	prefixes := make([]string, 0, len(rs))
	for prefix := range rs {
		prefixes = append(prefixes, string(prefix))
	}
	sort.Strings(prefixes)
	// Yes there's probably some algorithm to make this even more
	// efficient, but for now we're only dealing with a very small
	// number of repos, so less code and simpler code wins.
	for idx, prefix := range prefixes {
		for _, later := range prefixes[idx+1:] {
			if strings.HasPrefix(later, prefix) {
				return fmt.Errorf("Validation error: Two repositories found where one is a prefix of the other. This is not allowed, sorry. '%s' is a prefix of '%s'", prefix, later)
			}
		}
	}

	for _, repo := range rs {
		if err := repo.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) validate() error {
	if r == nil {
		return errors.New("Validation error: Repository can not be nil")
	} else if info, err := os.Stat(filepath.FromSlash(r.Directory)); err != nil {
		return err
	} else if !info.Mode().IsDir() {
		return fmt.Errorf("Validation error: Repository Directory is not a directory: '%s'", r.Directory)
	} else if err := r.maybeResolveBranch(); err != nil {
		return err
	} else if _, err := r.maybeResolveCommit(); err != nil {
		return err
	} else {
		return nil
	}
}

func (es Elements) validate(wf *Workflow) error {
	return utils.ErrorSlice{
		es.Types.validate(wf),
		es.Instances.validate(wf),
		es.InstancesConnections.validate(wf),
	}.Pack()
}

func (ets ElementTypes) validate(wf *Workflow) error {
	// we don't support import aliasing for elements. This means that
	// we require that every element type has a unique type name.
	namesToPath := make(map[ElementTypeName]ElementPath, len(ets))
	for _, et := range ets {
		if err := et.validate(wf); err != nil {
			return err
		} else if ep, found := namesToPath[et.Name()]; found {
			return fmt.Errorf("Validation error: ElementType '%v' is ambiguous (ElementPaths '%v' and '%v')", et.Name(), et.ElementPath, ep)
		} else if prefix := matchingPrefix(string(et.ElementPath), ".", "/"); prefix != "" {
			return fmt.Errorf("Validation error: Element Type %v: Element Path may not start with %v : %v", et.Name(), prefix, et.ElementPath)
		} else if substr := matchingContains(string(et.ElementPath), ".."); substr != "" {
			return fmt.Errorf("Validation error: Element Type %v: Element Path may not contain  %v : %v", et.Name(), substr, et.ElementPath)
		} else {
			namesToPath[et.Name()] = et.ElementPath
		}
	}
	return nil
}

func (et *ElementType) validate(wf *Workflow) error {
	if et == nil {
		return errors.New("Validation error: ElementType cannot be nil")
	} else if _, found := wf.Repositories[et.RepositoryName]; !found {
		return fmt.Errorf("Validation error: ElementType uses unknown RepositoryName: '%s'", et.RepositoryName)
	} else {
		return nil
	}
}

func (eis ElementInstances) validate(wf *Workflow) error {
	for name, ei := range eis {
		if name == "" {
			return errors.New("Validation error: Element Instance cannot have an empty name")
		} else if prefix := matchingPrefix(string(name), "."); prefix != "" {
			return fmt.Errorf("Validation error: Element Instance %v: name may not start with %v", name, prefix)
		} else if substr := matchingContains(string(name), "/", ".."); substr != "" {
			return fmt.Errorf("Validation error: Element Instance %v: name may not contain %v", name, substr)
		} else if err := ei.validate(wf, name); err != nil {
			return err
		}
	}
	return nil
}

func (ei *ElementInstance) validate(wf *Workflow, name ElementInstanceName) error {
	if ei == nil {
		return fmt.Errorf("Validation error: Element Instance %v cannot be nil", name)
	}
	tns := wf.TypeNames()
	if _, found := tns[ei.TypeName]; !found {
		maybeName := ElementType{ElementPath: ElementPath(ei.TypeName)}.Name()
		if _, found := tns[maybeName]; found {
			return fmt.Errorf("Validation error: Element Instance '%v' has unknown ElementTypeName '%v'. Did you mean '%v'?", name, ei.TypeName, maybeName)
		} else {
			return fmt.Errorf("Validation error: Element Instance '%v' has unknown ElementTypeName '%v'", name, ei.TypeName)
		}
	} else {
		ei.hasParameters = len(ei.Parameters) > 0
		return nil
	}
}

func (conns ElementInstancesConnections) validate(wf *Workflow) error {
	for _, conn := range conns {
		if err := conn.Source.validate(wf); err != nil {
			return err
		} else if err := conn.Target.validate(wf); err != nil {
			return err
		}
	}
	return nil
}

func (soc ElementSocket) validate(wf *Workflow) error {
	if ei, found := wf.Elements.Instances[soc.ElementInstance]; !found {
		return fmt.Errorf("Validation error: ElementConnection uses ElementInstance '%v' which does not exist.", soc.ElementInstance)
	} else if soc.ParameterName == "" {
		return fmt.Errorf("Validation error: ElementConnection using ElementInstance '%v' must specify a ParameterName.", soc.ElementInstance)
	} else {
		ei.hasConnections = true
		return nil
	}
}

func (inv Inventory) validate() error {
	return inv.PlateTypes.Validate()
}

func (cfg Config) validate() error {
	// NB the validation here is purely static - i.e. we're not
	// attempting to connect to any device plugins at this stage.
	seen := make(DeviceInstanceIDSet)
	return utils.ErrorSlice{
		cfg.GlobalMixer.validate(),
		cfg.GilsonPipetMax.validate(seen),
		cfg.Tecan.validate(seen),
		cfg.CyBio.validate(seen),
		cfg.Labcyte.validate(seen),
		cfg.Hamilton.validate(seen),
		cfg.QPCR.validate(seen),
		cfg.ShakerIncubator.validate(seen),
		cfg.PlateReader.validate(seen),
		cfg.assertOnlyOneMixer(),
	}.Pack()
}

func (cfg Config) assertOnlyOneMixer() error {
	if count := len(cfg.GilsonPipetMax.Devices) + len(cfg.Tecan.Devices) + len(cfg.CyBio.Devices) + len(cfg.Labcyte.Devices) + len(cfg.Hamilton.Devices); count > 1 {
		return fmt.Errorf("Currently a maximum of one mixer can be used per workflow. You have %d configured.", count)
	}
	return nil
}

type DeviceInstanceIDSet map[DeviceInstanceID]struct{}

func (dis DeviceInstanceIDSet) Add(id DeviceInstanceID) error {
	if _, found := dis[id]; found {
		return fmt.Errorf("Device IDs must be unique: multiple devices found with id %v", id)
	}
	dis[id] = struct{}{}
	return nil
}

func (global GlobalMixerConfig) validate() error {
	for idx, p := range global.InputPlates {
		if p == nil {
			return fmt.Errorf("GlobalMixer contains illegal nil input plate at index %d", idx)
		}
	}
	// We cannot validate plates and plate types until we have a
	// working inventory system.
	return nil
}

// Gilson
func (gilsons GilsonPipetMaxConfig) validate(seen DeviceInstanceIDSet) error {
	if err := gilsons.Defaults.validate("Defaults", true); err != nil {
		return err
	}
	for id, inst := range gilsons.Devices {
		if err := seen.Add(id); err != nil {
			return err
		} else if err := inst.validate(id, false); err != nil {
			return err
		}
	}
	return nil
}

func (inst *GilsonPipetMaxInstanceConfig) validate(id DeviceInstanceID, isDefaults bool) error {
	if len(id) == 0 {
		return errors.New("GilsonPipetMax: A device may not have an empty name.")

	} else if inst == nil {
		if isDefaults {
			return nil
		} else {
			return fmt.Errorf("GilsonPipetMax device '%s' has no configuration!", id)
		}

	} else if !isDefaults && strings.ToLower(string(id)) == "defaults" {
		return fmt.Errorf("Confusion: GilsonPipetMax device '%s' exists. Did you mean to set GilsonPipetMax.Defaults instead?", id)

	}
	return inst.CommonMixerInstanceConfig.validate(id)
}

// Tecan
func (tecans TecanConfig) validate(seen DeviceInstanceIDSet) error {
	if err := tecans.Defaults.validate("Defaults", true); err != nil {
		return err
	}
	for id, inst := range tecans.Devices {
		if err := seen.Add(id); err != nil {
			return err
		} else if err := inst.validate(id, false); err != nil {
			return err
		}
	}
	return nil
}

func (inst *TecanInstanceConfig) validate(id DeviceInstanceID, isDefaults bool) error {
	if len(id) == 0 {
		return errors.New("Tecan: A device may not have an empty name.")

	} else if inst == nil {
		if isDefaults {
			return nil
		} else {
			return fmt.Errorf("Tecan device '%s' has no configuration!", id)
		}

	} else if !isDefaults && strings.ToLower(string(id)) == "defaults" {
		return fmt.Errorf("Confusion: Tecan device '%s' exists. Did you mean to set Tecan.Defaults instead?", id)

	}
	return inst.CommonMixerInstanceConfig.validate(id)
}

// CyBio
func (cybios CyBioConfig) validate(seen DeviceInstanceIDSet) error {
	if err := cybios.Defaults.validate("Defaults", true); err != nil {
		return err
	}
	for id, inst := range cybios.Devices {
		if err := seen.Add(id); err != nil {
			return err
		} else if err := inst.validate(id, false); err != nil {
			return err
		}
	}
	return nil
}

func (inst *CyBioInstanceConfig) validate(id DeviceInstanceID, isDefaults bool) error {
	if len(id) == 0 {
		return errors.New("CyBio: A device may not have an empty name.")

	} else if inst == nil {
		if isDefaults {
			return nil
		} else {
			return fmt.Errorf("CyBio device '%s' has no configuration!", id)
		}

	} else if !isDefaults && strings.ToLower(string(id)) == "defaults" {
		return fmt.Errorf("Confusion: CyBio device '%s' exists. Did you mean to set CyBio.Defaults instead?", id)

	}
	return inst.CommonMixerInstanceConfig.validate(id)
}

// Labcyte
func (labcytes LabcyteConfig) validate(seen DeviceInstanceIDSet) error {
	if err := labcytes.Defaults.validate("Defaults", true); err != nil {
		return err
	}
	for id, inst := range labcytes.Devices {
		if err := seen.Add(id); err != nil {
			return err
		} else if err := inst.validate(id, false); err != nil {
			return err
		}
	}
	return nil
}

func (inst *LabcyteInstanceConfig) validate(id DeviceInstanceID, isDefaults bool) error {
	if len(id) == 0 {
		return errors.New("Labcyte: A device may not have an empty name.")

	} else if inst == nil {
		if isDefaults {
			return nil
		} else {
			return fmt.Errorf("Labcyte device '%s' has no configuration!", id)
		}

	} else if !isDefaults && strings.ToLower(string(id)) == "defaults" {
		return fmt.Errorf("Confusion: Labcyte device '%s' exists. Did you mean to set Labcyte.Defaults instead?", id)

	}
	// NB because the instruction plugin itself does validation of the model, we don't do that here!
	return inst.CommonMixerInstanceConfig.validate(id)
}

// Hamilton
func (hamiltons HamiltonConfig) validate(seen DeviceInstanceIDSet) error {
	if err := hamiltons.Defaults.validate("Defaults", true); err != nil {
		return err
	}
	for id, inst := range hamiltons.Devices {
		if err := seen.Add(id); err != nil {
			return err
		} else if err := inst.validate(id, false); err != nil {
			return err
		}
	}
	return nil
}

func (inst *HamiltonInstanceConfig) validate(id DeviceInstanceID, isDefaults bool) error {
	if len(id) == 0 {
		return errors.New("Hamilton: A device may not have an empty name.")

	} else if inst == nil {
		if isDefaults {
			return nil
		} else {
			return fmt.Errorf("Hamilton device '%s' has no configuration!", id)
		}

	} else if !isDefaults && strings.ToLower(string(id)) == "defaults" {
		return fmt.Errorf("Confusion: Hamilton device '%s' exists. Did you mean to set Hamilton.Defaults instead?", id)

	}
	// NB because the instruction plugin itself does validation of the model, we don't do that here!
	return inst.CommonMixerInstanceConfig.validate(id)
}

func (inst *CommonMixerInstanceConfig) validate(id DeviceInstanceID) error {
	if inst.ExecFile != "" {
		if abs, err := exec.LookPath(inst.ExecFile); err != nil {
			return fmt.Errorf("Error when trying to locate executable at %v for %v: %v", inst.ExecFile, id, err)
		} else {
			inst.ExecFile = abs
		}
	}
	// We cannot validate plates or tipes at this point because the
	// inventory may not be loaded. So those get validated later on.
	return inst.LayoutPreferences.validate()
}

func (lo *LayoutOpt) validate() error {
	if lo == nil {
		return nil
	}
	return utils.ErrorSlice{
		lo.Tipboxes.validate("Tipboxes"),
		lo.Inputs.validate("Inputs"),
		lo.Outputs.validate("Outputs"),
		lo.Tipwastes.validate("Tipwastes"),
		lo.Wastes.validate("Wastes"),
		lo.Washes.validate("Washes"),
	}.Pack()
}

func (a Addresses) validate(layoutOptionName string) error {
	if len(a.Map()) != len(a) {
		return fmt.Errorf("Layout option field %s has duplicate addresses: %v", layoutOptionName, a)
	}
	return nil
}

func (qpcr QPCRConfig) validate(seen DeviceInstanceIDSet) error {
	for id := range qpcr.Devices {
		if err := seen.Add(id); err != nil {
			return err
		}
	}
	return nil
}

func (si ShakerIncubatorConfig) validate(seen DeviceInstanceIDSet) error {
	for id := range si.Devices {
		if err := seen.Add(id); err != nil {
			return err
		}
	}
	return nil
}

func (pr PlateReaderConfig) validate(seen DeviceInstanceIDSet) error {
	for id := range pr.Devices {
		if err := seen.Add(id); err != nil {
			return err
		}
	}
	return nil
}

func matchingContains(str string, substrs ...string) string {
	for _, substr := range substrs {
		if strings.Contains(str, substr) {
			return substr
		}
	}
	return ""
}

func matchingPrefix(str string, prefixes ...string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(str, prefix) {
			return prefix
		}
	}
	return ""
}

func (sim *Simulation) validate(wf *Workflow) error {
	if sim == nil {
		return nil
	}
	// if there are simulated elements then we must have a simulation
	// id. But this is not an iff, because we could have a simulation
	// of no element instances...
	if elemCount := len(sim.Elements.Instances); elemCount != 0 && sim.SimulationId == "" {
		return fmt.Errorf("Validation error: Simulation records %d simulated elements, but we have no SimulationId", elemCount)
	} else if err := sim.SimulationId.Validate(elemCount == 0); err != nil {
		return err
	} else {
		return sim.Elements.validate(wf)
	}
}

func (sim *SimulatedElements) validate(wf *Workflow) error {
	return utils.ErrorSlice{
		sim.Types.validate(wf),
		sim.Instances.validate(wf),
	}.Pack()
}

func (types SimulatedElementTypes) validate(wf *Workflow) error {
	for name, elemTyp := range types {
		if _, found := wf.Repositories[elemTyp.RepositoryName]; !found {
			return fmt.Errorf("Validation error: Simulation records use element type %v with repository %v which is unknown",
				name, elemTyp.RepositoryName)
		}
	}
	return nil
}

func (insts SimulatedElementInstances) validate(wf *Workflow) error {
	tns := wf.TypeNames()
	for elemInstId, elemInst := range insts {
		if elemInst.ParentId == "" {
			// If there's no parent, then this is a top level
			// element. Which means it must be declared directly in
			// the workflow.
			if _, found := tns[elemInst.TypeName]; !found {
				return fmt.Errorf("Validation error: Simulation records top-level element instance (id: %v) with type %v, but that type is unknown in the workflow.",
					elemInstId, elemInst.TypeName)
			}
			if _, found := wf.Elements.Instances[elemInst.Name]; !found {
				return fmt.Errorf("Validation error: Simulation records top-level element instance (id: %v) with name %v, but that name is unknown in the workflow.",
					elemInst, elemInst.Name)
			}

		} else if _, found := insts[elemInst.ParentId]; !found {
			return fmt.Errorf("Validation error: Simulation records element instance (id: %v) with parent (id: %v), but the parent doesn't seem to exist",
				elemInstId, elemInst.ParentId)
		}

		if _, found := wf.Simulation.Elements.Types[elemInst.TypeName]; !found {
			return fmt.Errorf("Validation error: Simulation records element instance (id: %v) with type %v, but that type is unknown in the simulation types.",
				elemInstId, elemInst.TypeName)
		}
	}
	return nil
}
