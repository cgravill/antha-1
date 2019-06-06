package workflow

import (
	"encoding/json"
	"io"
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/utils"
)

// Testing contains data and configuration for testing.
type Testing struct {
	MixTaskChecks []MixTaskCheck `json:",omitempty"`
}

// MixTaskCheck contains check data for a single mix step.
type MixTaskCheck struct {
	Instructions json.RawMessage
	Outputs      map[string]*wtype.Plate
	TimeEstimate time.Duration
}

// WriteToFile writes the contents of a testing element to file as json.
func (t Testing) WriteToFile(filename string) error {
	if fh, err := utils.CreateFile(filename, utils.ReadWrite); err != nil {
		return err
	} else {
		defer fh.Close()
		return t.ToWriter(fh)
	}
}

// ToWriter writes a testing element as json.
func (t Testing) ToWriter(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(t)
}
