package main

import (
	"flag"

	"github.com/antha-lang/antha/logger"
	"github.com/antha-lang/antha/workflow"
)

func main() {
	flag.Usage = workflow.NewFlagUsage(nil,
		"Parse and transform go coverage profile data into coveralls format",
		"[flags] path/to/cover.profile",
		"github.com/antha-lang/antha/cmd/coveralls")

	var repoToken, commitSHA, repoName string
	flag.StringVar(&repoToken, "repotoken", "", "RepoToken for coveralls.")
	flag.StringVar(&commitSHA, "commitsha", "", "Git Commit SHA")
	flag.StringVar(&repoName, "reponame", "", "Name of git repository")
	flag.Parse()

	args := flag.Args()

	l := logger.NewLogger()

	pkgs := NewPackages()
	if err := pkgs.FromCoverProfile(args...); err != nil {
		logger.Fatal(l, err)
	}

	job := &Job{
		RepoToken:   repoToken,
		ServiceName: "antha",
		CommitSHA:   commitSHA,
		SourceFiles: pkgs.ToSourceFiles(repoName),
	}
	if err := job.Upload(); err != nil {
		logger.Fatal(l, err)
	}
}
