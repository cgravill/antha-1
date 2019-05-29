package composer

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"text/template"

	"github.com/antha-lang/antha/workflow"
)

// Because we generate code, and that code has a dependency on antha
// itself, we need to control exactly which versions of antha the
// generated code uses. Go mod allows us to do this, but it's worth
// keeping in mind that the tool binaries that we produce (composer
// etc) essentially are not stand-alone tools at all: they cannot be
// distributed or used without the source to antha.
//
// We need to be careful that in a developer scenario, with local
// checkouts of antha, we generate go.mod files that include replace
// directives to point back to the local checkout. That means that we
// need to determine whether we (i.e. this actual file) was compiled
// from within a local checkout, or within the GOPATH directory
// structure. This is ... just about possible.
//
// Also note that we simply do not support the old GOPATH mechanism
// any more.

var modtpl = `
{{define "repository"}}module {{.Name}}

require github.com/antha-lang/antha {{.AnthaVersion}}
{{end}}

{{define "workflow"}}module workflow

require (
	github.com/antha-lang/antha {{.AnthaVersion}}
{{range $repoName, $repo := .Repositories}}	{{$repoName}} v0.0.0
{{end}})
{{if .ReplaceAntha}}replace github.com/antha-lang/antha => {{.AnthaDir}}{{end}}
{{if .ReplaceRunner}}replace github.com/Synthace/antha-runner => {{.RunnerDir}}{{end}}
{{if .ReplacePlugins}}replace github.com/Synthace/instruction-plugins => {{.PluginsDir}}{{end}}
{{range $repoName, $repo := .Repositories}}replace {{$repoName}} => {{repopath $repoName}}
{{end}}{{end}}
`

type repositoryMod struct {
	Name         string
	AnthaVersion string
}

func newRepositoryMod(name string) *repositoryMod {
	rm := &repositoryMod{
		Name:         name,
		AnthaVersion: "v0.0.0",
	}

	if anthaMod := AnthaModule(); anthaMod != nil {
		if v := anthaMod.Version; len(v) > 0 && v[0] == 'v' {
			rm.AnthaVersion = v
		}
	}

	return rm
}

func AnthaModule() *debug.Module {
	if info, ok := debug.ReadBuildInfo(); ok {
		var anthaMod *debug.Module
		if info.Main.Path == "github.com/antha-lang/antha" {
			anthaMod = &info.Main
		} else {
			for _, mod := range info.Deps {
				if mod.Path == "github.com/antha-lang/antha" {
					anthaMod = mod
					break
				}
			}
		}
		if anthaMod == nil {
			return nil
		} else if anthaMod.Replace != nil {
			return anthaMod.Replace
		} else {
			return anthaMod
		}
	} else {
		return nil
	}
}

func repopath(repoName workflow.RepositoryName) string {
	return filepath.FromSlash(path.Join("../src", string(repoName)))
}

func renderRepositoryMod(w io.Writer, repoName workflow.RepositoryName) error {
	funcs := template.FuncMap{"repopath": repopath}
	if t, err := template.New("modtpl").Funcs(funcs).Parse(modtpl); err != nil {
		return err
	} else {
		return t.ExecuteTemplate(w, "repository", newRepositoryMod(string(repoName)))
	}
}

type workflowMod struct {
	*repositoryMod
	Repositories   workflow.Repositories
	AnthaDir       string
	ReplaceAntha   bool
	RunnerDir      string
	ReplaceRunner  bool
	PluginsDir     string
	ReplacePlugins bool
}

func newWorkflowMod(repos workflow.Repositories) *workflowMod {
	wm := &workflowMod{
		repositoryMod: newRepositoryMod(""),
		Repositories:  repos,
	}

	if wm.AnthaVersion == "v0.0.0" {
		wm.ReplaceAntha = true
		if _, file, _, ok := runtime.Caller(0); ok {
			wm.AnthaDir = filepath.Dir(filepath.Dir(file))
		}
		runnerDir := filepath.Join(filepath.Dir(wm.AnthaDir), "antha-runner")
		if _, err := os.Stat(runnerDir); !os.IsNotExist(err) {
			wm.ReplaceRunner = true
			wm.RunnerDir = runnerDir
		}
		pluginsDir := filepath.Join(filepath.Dir(wm.AnthaDir), "instruction-plugins")
		if _, err := os.Stat(pluginsDir); !os.IsNotExist(err) {
			wm.ReplacePlugins = true
			wm.PluginsDir = pluginsDir
		}
	}

	return wm
}

func renderWorkflowMod(w io.Writer, repos workflow.Repositories) error {
	funcs := template.FuncMap{"repopath": repopath}
	if t, err := template.New("modtpl").Funcs(funcs).Parse(modtpl); err != nil {
		return err
	} else {
		return t.ExecuteTemplate(w, "workflow", newWorkflowMod(repos))
	}
}
