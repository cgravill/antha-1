package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
	"github.com/antha-lang/antha/workflow/v1point2"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "All further args are interpreted as paths to up to date workflows to be merged and used as a basis for the upgrade.\n")
	}

	var from, outfile string
	flag.StringVar(&outfile, "outfile", "", "File to write to (default: will write to stdout)")
	flag.StringVar(&from, "from", "", "File to migrate (default: will be read from stdin)")
	flag.Parse()

	logger := logger.NewLogger()

	if source, err := workflow.ReaderFromPath(from); err != nil {
		logger.Fatal(err)
	} else if m, err := v1point2.NewMigrater(logger, flag.Args(), source); err != nil {
		logger.Fatal(err)
	} else if err := m.MigrateAll(); err != nil {
		logger.Fatal(err)
	} else {
		m.WriteTo(outfile)
	}
}
