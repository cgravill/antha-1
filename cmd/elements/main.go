package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
)

func main() {
	l := logger.NewLogger()

	args := os.Args
	if len(args) < 2 {
		l.Fatal(errors.New("Subcommand needed"))
	}

	subCmds := map[string]func(*logger.Logger, []string){
		"find":          find,
		"makeWorkflows": makeWorkflows,
	}

	if cmd, found := subCmds[args[1]]; found {
		cmd(l, args[2:])
	} else {
		l.Fatal(fmt.Errorf("Unknown subcommand: %s", args[1]))
	}
}

func find(l *logger.Logger, paths []string) {
	findElements(l, paths, func(f *workflow.File) error {
		l.Log("element", filepath.Dir(f.Name))
		return nil
	})
}

func makeWorkflows(l *logger.Logger, args []string) {
	outdir := ""
	flagset := flag.NewFlagSet("makeWorkflows", flag.ContinueOnError)
	flagset.StringVar(&outdir, "outdir", "", "Directory to write to")
	if err := flagset.Parse(args); err != nil {
		l.Fatal(err)
	}
	paths := flagset.Args()
	if err := os.MkdirAll(outdir, 0700); err != nil {
		l.Fatal(err)
	}
	findElements(l, paths, func(f *workflow.File) error {
		wf := &workflow.Workflow{
			JobId: workflow.JobId(filepath.Base(f.Name)),
			Elements: workflow.Elements{
				Types: []*workflow.ElementType{
					{
						ElementPath: workflow.ElementPath(filepath.Base(f.Name)),
					},
				},
			},
		}
		dir := filepath.Join(outdir, filepath.Dir(f.Name))
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		} else {
			return wf.WriteToFile(filepath.Join(dir, "workflow.json"))
		}
	})
}

func findElements(l *logger.Logger, paths []string, consumer func(*workflow.File) error) {
	if rs, err := workflow.ReadersFromPaths(paths); err != nil {
		l.Fatal(err)
	} else if wf, err := workflow.WorkflowFromReaders(rs...); err != nil {
		l.Fatal(err)
	} else {
		for _, r := range wf.Repositories {
			err := r.Walk(func(f *workflow.File) error {
				if f == nil || !f.IsRegular {
					return nil
				}
				if filepath.Ext(f.Name) != ".an" {
					return nil
				}
				return consumer(f)
			})
			if err != nil {
				l.Fatal(err)
			}
		}
	}
}