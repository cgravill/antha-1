package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
	"github.com/antha-lang/antha/workflow/v1_2"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "All further args are interpreted as paths to up to date workflows to be merged and used as a basis for the upgrade.\n")
	}

	var fromFile, toFile string
	flag.StringVar(&toFile, "to", "", "File to write to (default: will write to stdout)")
	flag.StringVar(&fromFile, "from", "", "File to migrate from (default: will be read from stdin)")
	flag.Parse()

	logger := logger.NewLogger()

	if source, err := workflow.ReaderFromPath(fromFile); err != nil {
		logger.Fatal(err)
	} else if m, err := v1_2.NewMigrater(logger, flag.Args(), source); err != nil {
		logger.Fatal(err)
	} else if err := m.MigrateAll(); err != nil {
		logger.Fatal(err)
	} else {
		m.WriteToPath(toFile)
	}
}
