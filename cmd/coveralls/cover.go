package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/antha-lang/antha/utils"
	"golang.org/x/tools/cover"
)

type Packages struct {
	m map[string]*Package
}

func NewPackages() *Packages {
	return &Packages{
		m: make(map[string]*Package),
	}
}

func (pkgs *Packages) FromCoverProfile(filenames ...string) error {
	for _, filename := range filenames {
		if profiles, err := cover.ParseProfiles(filename); err != nil {
			return err
		} else {
			for _, profile := range profiles {
				pkgs.AddProfile(profile)
			}
		}
	}
	return pkgs.LocateSources()
}

func (pkgs *Packages) AddProfile(profile *cover.Profile) {
	pkgName, fileName := filepath.Split(profile.FileName)
	pkg, found := pkgs.m[pkgName]
	if !found {
		pkg = &Package{
			Package: pkgName,
			Files:   make(map[string]*FileBlocks),
		}
		pkgs.m[pkgName] = pkg
	}
	pkg.AddProfile(fileName, profile)
}

// This must be called only after all profiles have been
// added. Sources are located in parallel: parallelism is at the
// package level; within a package, file sources are loaded
// sequentially. Provided there are plenty of packages to work with,
// you won't be able to go faster by adding concurrency at file level.
func (pkgs *Packages) LocateSources() error {
	concurrency := 10
	throttle := make(chan struct{}, concurrency)
	token := struct{}{}
	for c := 0; c < concurrency; c++ {
		throttle <- token
	}
	wg := new(sync.WaitGroup)
	wg.Add(len(pkgs.m))
	errors := make(utils.ErrorSlice, len(pkgs.m))
	errIdx := 0
	for _, pkg := range pkgs.m {
		go func(errIdx int, pkg *Package) {
			token := <-throttle

			errors[errIdx] = pkg.LocateSources()

			throttle <- token
			wg.Done()
		}(errIdx, pkg)
		errIdx++
	}
	wg.Wait()
	return errors.Pack()
}

type Package struct {
	Package string
	SrcDir  string
	Files   map[string]*FileBlocks
}

func (pkg *Package) LocateSources() error {
	if out, err := exec.Command("go", "list", "-f", "{{ .Dir }}", pkg.Package).Output(); err != nil {
		return err
	} else {
		dir := strings.TrimSuffix(string(out), "\n")
		if info, err := os.Stat(dir); err != nil {
			return err
		} else if !info.Mode().IsDir() {
			return fmt.Errorf("When resolving package %s: Not a directory: %s", pkg.Package, dir)
		} else {
			pkg.SrcDir = dir
			for _, blks := range pkg.Files {
				blks.ReadSource(pkg.SrcDir)
			}
			return nil
		}
	}
}

func (pkg *Package) AddProfile(fileName string, profile *cover.Profile) {
	if blks, found := pkg.Files[fileName]; !found {
		pkg.Files[fileName] = &FileBlocks{
			FileName: fileName,
			Blocks:   profile.Blocks,
		}
	} else {
		blks.Blocks = append(blks.Blocks, profile.Blocks...)
	}
}

type FileBlocks struct {
	FileName string
	Source   string
	Blocks   []cover.ProfileBlock
	Counts   []*int
}

func (blks *FileBlocks) ReadSource(dir string) error {
	if bs, err := ioutil.ReadFile(filepath.Join(dir, blks.FileName)); err != nil {
		return err
	} else {
		blks.Source = string(bs)
		// if a file has 1 line, it'll have 0 \n, hence the 1+
		counts := make([]*int, 1+strings.Count(blks.Source, "\n"))
		for _, blk := range blks.Blocks {
			for lineNo := blk.StartLine; lineNo <= blk.EndLine; lineNo++ {
				if lineNo >= len(counts) {
					return fmt.Errorf("File %s/%s has %d lines, but cover profile suggests line %d was hit!",
						dir, blks.FileName, len(counts), lineNo)
				}
				if countPtr := counts[lineNo]; countPtr == nil {
					num := blk.Count // explicitly copy the number because we may mutate it directly later
					counts[lineNo] = &num
				} else {
					*countPtr += blk.Count
				}
			}
		}
		blks.Counts = counts
		return nil
	}
}
