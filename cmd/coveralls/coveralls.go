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
	"strings"
)

type Job struct {
	RepoToken   string        `json:"repo_token"`
	ServiceName string        `json:"service_name"`
	SourceFiles []*SourceFile `json:"source_files"`
	CommitSHA   string        `json:"commit_sha"`
}

type SourceFile struct {
	Name         string `json:"name"`          // relative to repo
	SourceDigest string `json:"source_digest"` // md5, presumably in hex
	Coverage     []*int `json:"coverage"`
	Source       string `json:"source"`
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

func (pkg *Package) ToSourceFiles(prefix string) []*SourceFile {
	res := make([]*SourceFile, 0, len(pkg.Files))
	for fileName, blks := range pkg.Files {
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
