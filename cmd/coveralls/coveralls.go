package main

import (
	"bytes"
	"crypto/md5" // nolint
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Code in here is to create the necessary payload and upload to coveralls.

type Job struct {
	RepoToken    string        `json:"repo_token"`
	ServiceName  string        `json:"service_name"`
	ServiceJobId string        `json:"service_job_id,omitempty"`
	SourceFiles  []*SourceFile `json:"source_files"`
	Git          Git           `json:"git"`
}

type SourceFile struct {
	Name         string `json:"name"`          // relative to repo
	SourceDigest string `json:"source_digest"` // md5, presumably in hex
	Coverage     []*int `json:"coverage"`
	Source       string `json:"source"`
}

type Git struct {
	Head   Head   `json:"head"`
	Branch string `json:"branch,omitempty"`
}

type Head struct {
	Id             string `json:"id"`
	AuthorName     string `json:"author_name,omitempty"`
	AuthorEmail    string `json:"author_email,omitempty"`
	CommitterName  string `json:"committer_name,omitempty"`
	CommitterEmail string `json:"committer_email,omitempty"`
	Message        string `json:"message,omitempty"`
}

func (pkgs *Packages) ToSourceFiles(repoPrefix string) []*SourceFile {
	repoPrefix = strings.TrimSuffix(repoPrefix, "/") + "/"
	res := []*SourceFile{}
	for pkgName, pkg := range pkgs.m {
		if !strings.HasPrefix(pkgName, repoPrefix) {
			continue
		}
		res = append(res, pkg.ToSourceFiles(strings.TrimPrefix(pkgName, repoPrefix))...)
	}
	return res
}

// https://github.com/golang/go/issues/13560
var isGeneratedRegex = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT.$`)

func (pkg *Package) ToSourceFiles(prefix string) []*SourceFile {
	res := make([]*SourceFile, 0, len(pkg.Files))
	for fileName, blks := range pkg.Files {
		// Do not measure code coverage for generated code:
		if isGeneratedRegex.MatchString(blks.Source) {
			continue
		}
		md5sum := md5.Sum([]byte(blks.Source)) // nolint
		res = append(res, &SourceFile{
			Name:         path.Join(filepath.ToSlash(prefix), fileName),
			Coverage:     blks.Counts,
			Source:       blks.Source,
			SourceDigest: hex.EncodeToString(md5sum[:]),
		})
	}
	return res
}

func (job *Job) Upload() error {
	buf := new(bytes.Buffer)
	buf.Grow(10 * 1024 * 1024)
	if err := json.NewEncoder(buf).Encode(job); err != nil {
		return err
	}

	res, err := http.PostForm("https://coveralls.io/api/v1/jobs", url.Values{"json": {buf.String()}})
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(res.Body); err != nil {
			return err
		} else {
			res.Body.Close() // nolint
			return fmt.Errorf("Error calling coveralls: response status code: %d; body: %s", res.StatusCode, string(body))
		}
	} else {
		return nil
	}
}
