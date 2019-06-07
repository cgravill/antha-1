package liquidhandling

import (
	"fmt"
	"math"
	"strings"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
)

// ScoreChannel score the channels in terms of its suitability to transfer a volume vol, where lower is better
// Properties of the score function are:
//   - score is undefined for channels which cannot transfer vol and false is returned
//   - channels which can transfer vol in fewer transfers score better than those which take more
//   - for the same number of transfers, the channel which transfers vol while being proportionally closest to its maximum scores better
func ScoreChannel(vol wunit.Volume, lhcp *wtype.LHChannelParameter) (float64, bool) {
	if vol.LessThan(lhcp.Minvol.MinusEpsilon()) {
		return 0.0, false
	}

	nTransfers := math.Ceil(wunit.MustDivideVolumes(vol, lhcp.Maxvol))

	// assume that the volume will be divided evenly between each transfer
	volPerTransfer := wunit.CopyVolume(vol)
	volPerTransfer.DivideBy(float64(nTransfers))

	return nTransfers - wunit.MustDivideVolumes(volPerTransfer, lhcp.Maxvol), true
}

func tipHeadCompatible(tip *wtype.LHTip, head *wtype.LHHead) bool {
	//v1 - tip range must be contained entirely within head range

	return !(tip.MinVol.LessThan(head.Params.Minvol) || tip.MaxVol.GreaterThan(head.Params.Maxvol))
}

// ChooseChannels decide which combination of head/adaptor/tiptype should be used to transfer the given set of
// volumes. Currently the same combination will be returned for each channel, except for channels where
// vol[index] is zero, which are left empty
// len(vols) should be <= the number of channels of at least one available head.
func ChooseChannels(vols []wunit.Volume, prms *LHProperties) ([]*wtype.LHChannelParameter, []wtype.TipType, error) {

	// limitation: we can't distinguish between tips with the same maximum volume,
	// so we forbid configurations which have that

	// hash tips together by maximum volume
	maxVols := make(map[float64][]*wtype.LHTip, len(prms.Tips))
	for _, tip := range prms.Tips {
		max := tip.MaxVol.MustInStringUnit("ul").RawValue()
		maxVols[max] = append(maxVols[max], tip)
	}
	// check if any hashed to the same thing
	collidingTips := make([]*wtype.LHTip, 0, len(prms.Tips))
	for _, tips := range maxVols {
		if len(tips) > 1 {
			collidingTips = append(collidingTips, tips...)
		}
	}
	if len(collidingTips) > 0 {
		tipNames := make([]string, len(collidingTips))
		for i, tip := range collidingTips {
			tipNames[i] = tip.GetName()
		}
		return nil, nil, fmt.Errorf("cannot handle multiple tip types with the same maximum volume: %s", strings.Join(tipNames, ", "))
	}

	var bestChannel *wtype.LHChannelParameter
	var bestTipType wtype.TipType
	bestScore := math.MaxFloat64

	for _, head := range prms.GetLoadedHeads() {
		if len(vols) > head.Params.Multi {
			continue
		}
	TIP:
		for _, tip := range prms.Tips {
			if !tipHeadCompatible(tip, head) {
				continue
			}
			score := 0.0
			for _, vol := range vols {
				if vol.IsZero() {
					continue
				}
				if s, ok := ScoreChannel(vol, head.Params.MergeWithTip(tip)); !ok {
					// this volume can't be handled by the tip, so move on to the next
					continue TIP
				} else {
					score += s
				}
			}

			if score < bestScore {
				bestScore = score
				bestChannel = head.Params
				bestTipType = tip.Type
			}
		}
	}

	if bestChannel == nil {
		var config []string
		heads := prms.GetLoadedHeads()
		for _, head := range heads {
			config = append(config, head.String())
			for _, tip := range prms.Tips {
				if tipHeadCompatible(tip, head) {
					config = append(config, "\t"+tip.String())
				}
			}
		}
		return nil, nil, wtype.LHErrorf(wtype.LH_ERR_VOL, "no tip chosen: volumes %v could not be moved by the liquid handler in this configuration\n\t%s", vols, strings.Join(config, "\n\t"))
	}

	props := make([]*wtype.LHChannelParameter, len(vols))
	tipTypes := make([]wtype.TipType, len(vols))
	for i, vol := range vols {
		if !vol.IsZero() {
			props[i] = bestChannel
			tipTypes[i] = bestTipType
		}
	}

	propNames := make([]string, len(props))
	for i, prop := range props {
		if prop != nil {
			propNames[i] = prop.Name
		}
	}

	return props, tipTypes, nil
}
