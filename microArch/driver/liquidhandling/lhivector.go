package liquidhandling

import (
	"fmt"
	"strings"

	"github.com/Synthace/antha/antha/anthalib/wtype"
	"github.com/Synthace/antha/laboratory/effects/id"
)

type LHIVector []*wtype.LHInstruction

func (lhiv LHIVector) String() string {
	ss := []string{}

	for _, v := range lhiv {
		ss = append(ss, v.String())
	}

	return strings.Join(ss, "\n")
}

func (lhiv LHIVector) MaxLen() int {
	l := 0
	for _, i := range lhiv {
		if i == nil {
			continue
		}
		ll := len(i.Inputs)

		if ll > l {
			l = ll
		}
	}

	return l
}

func (lhiv LHIVector) CompsAt(idGen *id.IDGenerator, i int) []*wtype.Liquid {
	ret := make([]*wtype.Liquid, len(lhiv))

	for ix, ins := range lhiv {
		if i == 0 && ins.IsMixInPlace() {
			continue
		}

		if ins == nil || i >= len(ins.Inputs) {
			continue
		}

		ret[ix] = ins.Inputs[i].Dup(idGen)

		ret[ix].Loc = ins.PlateID + ":" + ins.Welladdress
	}

	return ret
}

func (lhiv LHIVector) Generations() string {
	s := ""
	for _, i := range lhiv {
		if i == nil {
			continue
		}
		s += fmt.Sprintf("%d,", i.Generation())
	}
	return s
}
