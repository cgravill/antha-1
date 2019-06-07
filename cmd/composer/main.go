package main

import (
	"flag"
	"os/exec"
	"path/filepath"

	runner "github.com/Synthace/antha-runner/export"
	"github.com/Synthace/antha/composer"
	"github.com/Synthace/antha/logger"
	"github.com/Synthace/antha/workflow"
)

func main() {
	flag.Usage = workflow.NewFlagUsage(nil,
		"Parse, compile and run a workflow",
		"[flags] [workflow-snippet.json...]",
		"github.com/Synthace/antha/cmd/composer")

	var inDir, outDir string
	var keep, run, linkedDrivers bool
	flag.StringVar(&inDir, "indir", "", "Directory from which to read files (optional)")
	flag.StringVar(&outDir, "outdir", "", "Directory to write to (default: a temporary directory will be created)")
	flag.BoolVar(&keep, "keep", false, "Keep build environment if compilation is successful")
	flag.BoolVar(&run, "run", true, "Run the workflow if compilation is successful")
	flag.BoolVar(&linkedDrivers, "linkedDrivers", true, "Compile workflow with linked-in drivers")
	flag.Parse()

	l := logger.NewLogger()

	if err := compose(l, inDir, outDir, keep, run, linkedDrivers); err != nil {
		// if the workflow ran but failed then we trust it to export its
		// own errors. So we only need to export errors if:
		// - the err is not an ExitError (i.e. we didn't even run the workflow)
		// - or the exit code is unexpected (not 1 or 0)
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			if exportErr := runner.Export(nil, inDir, outDir, nil, err); exportErr != nil {
				l.Log("exportError", exportErr)
			}
		}
		logger.Fatal(l, err)
	} else {
		l.Log("progress", "complete")
	}
}

func compose(l *logger.Logger, inDir, outDir string, keep, run, linkedDrivers bool) error {
	if wfPaths, err := workflow.GatherPaths(nil, filepath.Join(inDir, "workflow")); err != nil {
		return err
	} else if rs, err := workflow.ReadersFromPaths(wfPaths); err != nil {
		return err
	} else if wf, err := workflow.WorkflowFromReaders(rs...); err != nil {
		return err
	} else if err := wf.Validate(); err != nil {
		return err
	} else if cb, err := composer.NewComposerBase(l, inDir, outDir); err != nil {
		return err
	} else {
		defer cb.CloseLogs()
		return cb.ComposeMainAndRun(keep, run, linkedDrivers, wf)
	}
}
