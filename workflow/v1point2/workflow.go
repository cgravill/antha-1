package v1point2

import (
	"encoding/json"
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
)

type Workflow struct {
	Desc
	TestOpt
	RawParams
	Version    string             `json:"version"`
	Properties workflowProperties `json:"Properties"`
}

type workflowProperties struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RawParams struct {
	Parameters map[string]map[string]json.RawMessage `json:"Parameters"`
	Config     *Opt                                  `json:"Config"`
}

type Desc struct {
	Processes   map[string]Process `json:"Processes"`
	Connections []Connection       `json:"connections"`
}

type Connection struct {
	Src Port `json:"source"`
	Tgt Port `json:"target"`
}

type Process struct {
	Component string         `json:"component"`
	Metadata  screenPosition `json:"metadata"`
}

type screenPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Port struct {
	Process string `json:"process"`
	Port    string `json:"port"`
}

type TestOpt struct {
	ComparisonOptions   string
	CompareInstructions bool
	CompareOutputs      bool
	Results             TestResults
}

type TestResults struct {
	MixTaskResults []MixTaskResult
}

type MixTaskResult struct {
	Instructions liquidhandling.SetOfRobotInstructions
	Outputs      map[string]*wtype.Plate
	TimeEstimate time.Duration
}
