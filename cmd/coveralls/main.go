package main

import (
	"flag"

	"github.com/Synthace/antha/logger"
	"github.com/Synthace/antha/workflow"
)

func main() {
	flag.Usage = workflow.NewFlagUsage(nil,
		"Parse and transform go coverage profile data into coveralls format",
		"[flags] path/to/cover.profile",
		"github.com/Synthace/antha/cmd/coveralls")

	var repoName, repoToken, commitSHA, branchName, buildId string
	flag.StringVar(&repoName, "reponame", "", "Name of git repository")
	flag.StringVar(&repoToken, "repotoken", "", "RepoToken for coveralls.")
	flag.StringVar(&commitSHA, "commitsha", "", "Git Commit SHA")
	flag.StringVar(&branchName, "branchname", "", "Git Branch Name")
	flag.StringVar(&buildId, "buildid", "", "Build Id")
	flag.Parse()

	args := flag.Args()

	l := logger.NewLogger()

	pkgs := NewPackages()
	if err := pkgs.FromCoverProfile(args...); err != nil {
		logger.Fatal(l, err)
	}

	job := &Job{
		RepoToken:    repoToken,
		ServiceName:  "antha-coveralls",
		ServiceJobId: buildId,
		SourceFiles:  pkgs.ToSourceFiles(repoName),
		Git: Git{
			Head: Head{
				Id: commitSHA,
			},
			Branch: branchName,
		},
	}
	if err := job.Upload(); err != nil {
		logger.Fatal(l, err)
	}
}
