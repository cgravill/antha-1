package mixer

import (
	"github.com/Synthace/antha/inventory"
	"github.com/Synthace/antha/microArch/scheduler/liquidhandling"
	"github.com/Synthace/antha/workflow"
)

type GlobalMixerConfig struct {
	*workflow.GlobalMixerConfig
}

func NewGlobalMixerConfig(inv *inventory.Inventory, cfg *workflow.GlobalMixerConfig) (*GlobalMixerConfig, error) {
	global := &GlobalMixerConfig{
		GlobalMixerConfig: cfg,
	}
	if err := global.Validate(inv); err != nil {
		return nil, err
	} else {
		return global, nil
	}
}

func (cfg *GlobalMixerConfig) Validate(inv *inventory.Inventory) error {
	for _, plate := range cfg.InputPlates {
		if _, err := inv.Plates.NewPlateType(plate.Type); err != nil {
			return err
		}
	}
	return nil
}

func (cfg *GlobalMixerConfig) ApplyToLHRequest(req *liquidhandling.LHRequest) error {
	if cfg.CustomPolicyRuleSet != nil {
		req.AddUserPolicies(cfg.CustomPolicyRuleSet)
	}
	if err := req.PolicyManager.SetOption("USE_DRIVER_TIP_TRACKING", cfg.UseDriverTipTracking); err != nil {
		return err
	}
	req.Options.PrintInstructions = cfg.PrintInstructions
	req.Options.IgnorePhysicalSimulation = cfg.IgnorePhysicalSimulation
	return nil
}
