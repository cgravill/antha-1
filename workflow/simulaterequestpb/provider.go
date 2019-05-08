package simulaterequestpb

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/Synthace/antha-runner/aconv"
	"github.com/Synthace/antha-runner/protobuf"
	inventorypb "github.com/Synthace/microservice/cmd/inventory/protobuf"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wtype/liquidtype"
	"github.com/antha-lang/antha/laboratory/effects"
	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
	"github.com/antha-lang/antha/workflow/migrate"
	"github.com/golang/protobuf/proto"
)

type Provider struct {
	pb         *protobuf.SimulateRequest
	labEffects *effects.LaboratoryEffects
	repoMap    workflow.ElementTypesByRepository
	logger     *logger.Logger
}

func NewProvider(
	inputReader io.Reader,
	fm *effects.FileManager,
	repoMap workflow.ElementTypesByRepository,
	gilsonDeviceName string,
	logger *logger.Logger,
) (*Provider, error) {
	bytes, err := ioutil.ReadAll(inputReader)
	if err != nil {
		return nil, err
	}

	pb := &protobuf.SimulateRequest{}
	if err := proto.Unmarshal(bytes, pb); err != nil {
		return nil, err
	}

	id, err := workflow.RandomBasicId("simulaterequestpb")
	if err != nil {
		return nil, err
	}

	return &Provider{
		pb:         pb,
		labEffects: effects.NewLaboratoryEffects(nil, id, fm),
		repoMap:    repoMap,
		logger:     logger,
	}, nil
}

func (p *Provider) GetWorkflowID() (workflow.BasicId, error) {
	id, err := workflow.RandomBasicId("")
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *Provider) GetMeta() (workflow.Meta, error) {
	// No-op for this provider type, it doesn't model metadata
	return workflow.Meta{}, nil
}

func (p *Provider) GetRepositories() (workflow.Repositories, error) {
	// No-op for this provider type, it doesn't model repositories
	return workflow.Repositories{}, nil
}

func (p *Provider) migrateElementParameters(fm *effects.FileManager, process *protobuf.Process) (workflow.ElementParameterSet, error) {
	pset := make(workflow.ElementParameterSet)

	for _, param := range process.Parameters {
		// Could be a raw param (JSON) or a reference (opaque string)
		data := param.GetRaw()
		if len(data) > 0 {
			// It's a raw param
			var rawJSON json.RawMessage
			err := json.Unmarshal(data, &rawJSON)
			if err != nil {
				return pset, err
			}
			pval, err := migrate.MaybeMigrateFileParam(fm, rawJSON)
			if err != nil {
				return pset, err
			}
			pset[workflow.ElementParameterName(param.GetName())] = pval
		} else {
			// It _could_ be a reference (param.GetReference() could be
			// non-empty), but we're pretty sure the reference stuff isn't used
			// (https://github.com/antha-lang/antha/pull/1068#discussion_r281625589),
			// so if there isn't a raw value, we return an error.
			return pset, fmt.Errorf("Param %v has no data", param.GetName())
		}
	}

	return pset, nil
}

func (p *Provider) migrateElement(fm *effects.FileManager, process *protobuf.Process) (*workflow.ElementInstance, error) {
	ei := &workflow.ElementInstance{}

	ei.ElementTypeName = workflow.ElementTypeName(process.Component)

	params, err := p.migrateElementParameters(fm, process)
	if err != nil {
		return nil, err
	}
	ei.Parameters = params
	return ei, nil
}

func (p *Provider) getElementInstances() (workflow.ElementInstances, error) {
	instances := workflow.ElementInstances{}
	for _, process := range p.pb.Processes {
		name := workflow.ElementInstanceName(process.Id)
		ei, err := p.migrateElement(p.labEffects.FileManager, process)
		if err != nil {
			return nil, err
		}
		instances[name] = ei
	}
	return instances, nil
}

func (p *Provider) getElementTypes() (workflow.ElementTypes, error) {
	seen := make(map[string]struct{}, len(p.pb.Processes))
	types := make(workflow.ElementTypes, 0, len(p.pb.Processes))
	for _, v := range p.pb.Processes {
		if _, found := seen[v.Component]; found {
			continue
		}

		seen[v.Component] = struct{}{}
		et, err := migrate.UniqueElementType(p.repoMap, workflow.ElementTypeName(v.Component))
		if err != nil {
			return nil, err
		}
		types = append(types, et)
	}

	return types, nil
}

func (p *Provider) getElementConnections() (workflow.ElementInstancesConnections, error) {
	connections := make(workflow.ElementInstancesConnections, 0, len(p.pb.Connections))
	for _, c := range p.pb.Connections {
		connections = append(connections, workflow.ElementConnection{
			Source: workflow.ElementSocket{
				ElementInstance: workflow.ElementInstanceName(c.Source.Process),
				ParameterName:   workflow.ElementParameterName(c.Source.Port),
			},
			Target: workflow.ElementSocket{
				ElementInstance: workflow.ElementInstanceName(c.Target.Process),
				ParameterName:   workflow.ElementParameterName(c.Target.Port),
			},
		})
	}
	return connections, nil
}

func (p *Provider) GetElements() (workflow.Elements, error) {
	instances, err := p.getElementInstances()
	if err != nil {
		return workflow.Elements{}, err
	}

	types, err := p.getElementTypes()
	if err != nil {
		return workflow.Elements{}, err
	}

	connections, err := p.getElementConnections()
	if err != nil {
		return workflow.Elements{}, err
	}

	return workflow.Elements{
		Instances:            instances,
		Types:                types,
		InstancesConnections: connections,
	}, nil
}

func (p *Provider) translatePlateTypes(plateTypes []*inventorypb.PlateType) (wtype.PlateTypes, error) {
	result := wtype.PlateTypes{}

	for _, plateTypePB := range plateTypes {
		plateType, err := aconv.LHPlateTypeFromProtobuf(p.labEffects.IDGenerator, plateTypePB)
		if err != nil {
			return nil, err
		}
		result[plateType.Name] = plateType
	}

	p.labEffects.Inventory.Plates.SetPlateTypes(result)

	return result, nil
}

// as a side effect, this populates the provider's inventory, so this
// MUST be called before we translate the mixer config (which has
// input plates, which will rely on xref into the inventory)
func (p *Provider) GetInventory() (workflow.Inventory, error) {
	plateTypes, err := p.translatePlateTypes(p.pb.GetPlateTypes())
	if err != nil {
		return workflow.Inventory{}, err
	}
	return workflow.Inventory{
		PlateTypes: plateTypes,
	}, nil
}

func (p *Provider) translatePlates(plates []*inventorypb.Plate) ([]*wtype.Plate, error) {
	result := make([]*wtype.Plate, len(plates))
	for i, platePB := range plates {
		plate, err := aconv.LHPlateFromProtobuf(p.labEffects, platePB)
		if err != nil {
			return nil, err
		}
		result[i] = plate
	}
	return result, nil
}

func (p *Provider) getGlobalMixerConfig() (workflow.GlobalMixerConfig, error) {
	config := workflow.GlobalMixerConfig{}
	mc := p.pb.GetMixerConfig()
	if mc != nil {

		if mc.LiquidHandlingPolicyXlsxJmpFile != nil {
			policyMap, err := liquidtype.PolicyMakerFromBytes(mc.LiquidHandlingPolicyXlsxJmpFile, wtype.PolicyName(liquidtype.BASEPolicy))
			if err != nil {
				return workflow.GlobalMixerConfig{}, err
			}
			lhpr := wtype.NewLHPolicyRuleSet()
			lhpr, err = wtype.AddUniversalRules(lhpr, policyMap)
			if err != nil {
				return workflow.GlobalMixerConfig{}, err
			}
			config.CustomPolicyRuleSet = lhpr
		}

		config.IgnorePhysicalSimulation = mc.IgnorePhysicalSimulation
		if inputPlates, err := p.translatePlates(mc.GetInputPlateVals().GetPlates()); err != nil {
			return workflow.GlobalMixerConfig{}, err
		} else {
			config.InputPlates = inputPlates
		}
		config.UseDriverTipTracking = mc.UseDriverTipTracking
	}
	return config, nil
}

func (p *Provider) getGilsonPipetMaxInstanceConfig() (*workflow.GilsonPipetMaxInstanceConfig, error) {
	config := workflow.GilsonPipetMaxInstanceConfig{}
	mc := p.pb.GetMixerConfig()
	if mc != nil {
		config.InputPlateTypes = migrate.UpdatePlateTypes(mc.InputPlateTypes)
		config.MaxPlates = &mc.MaxPlates
		config.MaxWells = &mc.MaxWells
		config.OutputPlateTypes = migrate.UpdatePlateTypes(mc.OutputPlateTypes)
		config.ResidualVolumeWeight = &mc.ResidualVolumeWeight
		config.TipTypes = mc.TipTypes
		config.LayoutPreferences = &workflow.LayoutOpt{
			Inputs:    mc.DriverSpecificInputPreferences,
			Outputs:   mc.DriverSpecificOutputPreferences,
			Tipboxes:  mc.DriverSpecificTipPreferences,
			Tipwastes: mc.DriverSpecificTipWastePreferences,
			Washes:    mc.DriverSpecificWashPreferences,
		}
	}
	return &config, nil
}

func (p *Provider) getGilsonPipetMaxConfig() (workflow.GilsonPipetMaxConfig, error) {
	devices := map[workflow.DeviceInstanceID]*workflow.GilsonPipetMaxInstanceConfig{}
	gilsonPipetMaxInstanceConfig, err := p.getGilsonPipetMaxInstanceConfig()
	if err != nil {
		return workflow.GilsonPipetMaxConfig{}, err
	}

	for _, device := range p.pb.GetAvailable() {
		class := device.GetClass()
		switch class {
		case "GilsonPipetMax":
			devices[workflow.DeviceInstanceID(device.GetId())] = gilsonPipetMaxInstanceConfig
		default:
			return workflow.GilsonPipetMaxConfig{}, fmt.Errorf("Unsupported device class: %v", class)
		}
	}

	return workflow.GilsonPipetMaxConfig{
		Devices: devices,
	}, nil
}

func (p *Provider) GetConfig() (workflow.Config, error) {
	gmc, err := p.getGlobalMixerConfig()
	if err != nil {
		return workflow.Config{}, err
	}

	gpmc, err := p.getGilsonPipetMaxConfig()
	if err != nil {
		return workflow.Config{}, err
	}

	return workflow.Config{
		GlobalMixer:    gmc,
		GilsonPipetMax: gpmc,
	}, nil
}

func (p *Provider) GetTesting() (workflow.Testing, error) {
	// No-op for this provider type, it doesn't model test results
	return workflow.Testing{}, nil
}
