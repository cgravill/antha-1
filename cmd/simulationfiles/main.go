package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/utils"
	"github.com/antha-lang/antha/workflow"
)

func main() {
	flag.Usage = workflow.NewFlagUsage(nil,
		"Export files created by a simulation",
		"[flags] [workflow-snippet.json...]",
		"github.com/antha-lang/antha/cmd/simulationfiles")

	var inDir, outDir string
	flag.StringVar(&inDir, "indir", "", "Directory from which to read files (optional)")
	flag.StringVar(&outDir, "outdir", "", "Directory to write files to (optional)")
	flag.Parse()

	l := logger.NewLogger()
	if outDir == "" {
		if d, err := ioutil.TempDir("", "antha-simulationfiles"); err != nil {
			logger.Fatal(l, err)
			return
		} else {
			l.Log("outdir", d)
			outDir = d
		}
	}

	if err := extract(l, inDir, outDir); err != nil {
		logger.Fatal(l, err)
	}
}

func extract(l *logger.Logger, inDir, outDir string) error {
	if wfPaths, err := workflow.GatherPaths(nil, inDir); err != nil {
		return err
	} else if rs, err := workflow.ReadersFromPaths(wfPaths); err != nil {
		return err
	} else if wf, err := workflow.WorkflowFromReaders(rs...); err != nil {
		return err
	} else if err := wf.Validate(); err != nil {
		return err
	} else {

		if wf.Simulation == nil {
			return errors.New("Workflow does not contain any record of simulation")
		}
		sim := wf.Simulation
		fmt.Printf("Summary:\n Workflow Id:      %v\n Simulation Id:    %v\n Antha Version:    %v\n Simulation Start: %v\n Simulation End:   %v\n",
			wf.WorkflowId, sim.SimulationId, sim.Version, sim.Start, sim.End)

		elemTypes := sim.Elements.Types
		for id, inst := range sim.Elements.Instances {
			instDir := filepath.Join(outDir, fmt.Sprintf("%s_%s", id, inst.Name))
			elemType := elemTypes[inst.TypeName]
			if err := extractFields(l, sim, instDir, "output", elemType.OutputsFieldTypes, id, inst); err != nil {
				return err
			}
			if err := extractFields(l, sim, instDir, "data", elemType.DataFieldTypes, id, inst); err != nil {
				return err
			}
		}
		return nil
	}
}

func extractFields(l *logger.Logger, sim *workflow.Simulation, outDir string, fieldGroup string, fields map[workflow.ElementParameterName]string, id workflow.ElementInstanceId, inst workflow.SimulatedElementInstance) error {
	for paramName, paramType := range fields {
		dir := filepath.Join(outDir, string(paramName))
		if paramType == "*github.com/antha-lang/antha/antha/anthalib/wtype.File" {
			l.Log("id", id, "type", inst.TypeName, "name", inst.Name, fieldGroup, paramName, "dir", dir)
			if err := writeFile(l, sim, dir, inst.Files[paramName], ""); err != nil {
				return err
			}
		} else if paramType == "[]*github.com/antha-lang/antha/antha/anthalib/wtype.File" {
			l.Log("id", id, "type", inst.TypeName, "name", inst.Name, fieldGroup, paramName, "dir", dir)
			if err := writeFiles(l, sim, dir, inst.Files[paramName]); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(l *logger.Logger, sim *workflow.Simulation, dir string, bs json.RawMessage, prefix string) error {
	f := wtype.File{}
	if err := json.Unmarshal(bs, &f); err != nil {
		return err
	}
	dst := filepath.Join(dir, prefix+f.Name)
	src := filepath.Join(sim.InDir, "data", f.Path())
	if f.IsOutput() {
		src = filepath.Join(sim.OutDir, "data", f.Path())
	}
	l.Log("src", src, "dst", dst)
	if err := utils.MkdirAll(dir); err != nil {
		return err
	}
	if srcFH, err := os.Open(src); err != nil {
		return err
	} else {
		defer srcFH.Close()
		if dstFH, err := utils.CreateFile(dst, utils.ReadWrite); err != nil {
			return err
		} else {
			defer dstFH.Close()
			_, err := io.Copy(dstFH, srcFH)
			return err
		}
	}
}

func writeFiles(l *logger.Logger, sim *workflow.Simulation, dir string, bs json.RawMessage) error {
	fs := []json.RawMessage{}
	if err := json.Unmarshal(bs, &fs); err != nil {
		return err
	}
	for idx, f := range fs {
		if err := writeFile(l, sim, dir, f, fmt.Sprintf("%d_", idx)); err != nil {
			return err
		}
	}
	return nil
}
