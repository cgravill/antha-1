package liquidhandling

import (
	"fmt"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"math"
)

// TransferVolumes returns a slice of volumes V such that Min <= v <= Max and sum(V) = Vol
func TransferVolumes(vol, min, max wunit.Volume) ([]wunit.Volume, error) {

	if vol.LessThan(min.MinusEpsilon()) {
		return nil, wtype.LHError(wtype.LH_ERR_VOL, fmt.Sprintf("Liquid Handler cannot service volume %s requested: minimum volume is %s", vol, min))
	}

	// how many transfers should we make?
	// Epsilon() here is used to avoid floating point errors e.g. 100+1e-8 ul taking 3 transfers with a 50ul tip
	n := math.Ceil(wunit.MustDivide(vol.MinusEpsilon(), max.PlusEpsilon()))
	transferVol := wunit.CopyVolume(vol)
	transferVol.DivideBy(n)

	ret := make([]wunit.Volume, int(n))
	for i := range ret {
		ret[i] = wunit.CopyVolume(transferVol)
	}

	return ret, nil
}

// TransferVolumesMulti given a slice of volumes to transfer and channels to use,
// return an array of transfers to make such that `ret[i][j]` is the volume of the ith transfer to be made with channel j
func TransferVolumesMulti(vols VolumeSet, chans []*wtype.LHChannelParameter) ([]VolumeSet, error) {

	// aggregate vertically
	mods := make([]VolumeSet, len(vols))
	mx := 0
	for i := 0; i < len(vols); i++ {
		if chans[i] == nil {
			continue
		}
		mod, err := TransferVolumes(vols[i], chans[i].Minvol, chans[i].Maxvol)

		if err != nil {
			return []VolumeSet{}, err
		}

		mods[i] = mod
		if len(mod) > mx {
			mx = len(mod)
		}

	}

	ret := make([]VolumeSet, mx)

	for j := 0; j < mx; j++ {
		vs := make(VolumeSet, len(vols))

		for i := 0; i < len(vols); i++ {
			if j < len(mods[i]) {
				vs[i] = mods[i][j]
			}
		}

		ret[j] = vs
	}

	return ret, nil
}
