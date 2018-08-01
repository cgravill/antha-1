package wtype

// adaptor
type LHAdaptor struct {
	Name         string
	ID           string
	Manufacturer string
	Params       *LHChannelParameter
	Tips         []*LHTip
}

//NewLHAdaptor make a new adaptor
func NewLHAdaptor(name, mf string, params *LHChannelParameter) *LHAdaptor {
	return &LHAdaptor{
		Name:         name,
		Manufacturer: mf,
		Params:       params,
		ID:           GetUUID(),
		Tips:         []*LHTip{},
	}
}

//Dup duplicate the adaptor and any loaded tips with new IDs
func (lha *LHAdaptor) Dup() *LHAdaptor {
	return lha.dup(false)
}

//AdaptorType the manufacturer and name of the adaptor
func (lha *LHAdaptor) AdaptorType() string {
	return lha.Manufacturer + lha.Name
}

//DupKeepIDs duplicate the adaptor and any loaded tips keeping the same IDs
func (lha *LHAdaptor) DupKeepIDs() *LHAdaptor {
	return lha.dup(true)
}

func (lha *LHAdaptor) dup(keepIDs bool) *LHAdaptor {
	var params *LHChannelParameter
	if keepIDs {
		params = lha.Params.DupKeepIDs()
	} else {
		params = lha.Params.Dup()
	}

	ad := NewLHAdaptor(lha.Name, lha.Manufacturer, params)

	for i, tip := range lha.Tips {
		if tip != nil {
			if keepIDs {
				ad.AddTip(i, tip.DupKeepID())
			} else {
				ad.AddTip(i, tip.Dup())
			}
		}
	}

	if keepIDs {
		ad.ID = lha.ID
	} else {
		ad.ID = GetUUID()
	}

	return ad
}

//NTipsLoaded the number of tips currently loaded
func (lha *LHAdaptor) NTipsLoaded() int {
	r := 0
	for i := range lha.Tips {
		if lha.Tips[i] != nil {
			r += 1
		}
	}
	return r
}

//IsTipLoaded Is there a tip loaded on channel_number
func (lha *LHAdaptor) IsTipLoaded(channel_number int) bool {
	return lha.Tips[channel_number] != nil
}

//GetTip Return the tip at channel_number, nil otherwise
func (lha *LHAdaptor) GetTip(channel_number int) *LHTip {
	return lha.Tips[channel_number]
}

//AddTip Load a tip to the specified channel
func (lha *LHAdaptor) AddTip(channel_number int, tip *LHTip) {
	lha.Tips[channel_number] = tip
}

//RemoveTip Remove a tip from the specified channel and return it
func (lha *LHAdaptor) RemoveTip(channel_number int) *LHTip {
	tip := lha.Tips[channel_number]
	lha.Tips[channel_number] = nil
	return tip
}

//RemoveTips Remove every tip from the adaptor
func (lha *LHAdaptor) RemoveTips() []*LHTip {
	ret := make([]*LHTip, 0, lha.NTipsLoaded())
	for i := range lha.Tips {
		if lha.Tips[i] != nil {
			ret = append(ret, lha.Tips[i])
			lha.Tips[i] = nil
		}
	}
	return ret
}

//GetParams get the channel parameters for the adaptor, combined with any loaded tips
func (lha *LHAdaptor) GetParams() *LHChannelParameter {
	if lha.NTipsLoaded() == 0 {
		return lha.Params
	} else {
		params := *lha.Params
		for _, tip := range lha.Tips {
			if tip != nil {
				params = *params.MergeWithTip(tip)
			}
		}
		return &params
	}
}