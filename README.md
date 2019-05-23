# Antha

For a quick guide walking through a installation and basic tasks, look at the [quickstart guide](README_Quickstart.md).
This document contains more detailed information, and discussion of further options for installation and use.

## Getting Started

There are a number of installation options, depending on your use case.

1. I am a user of Antha, not a developer of Antha; I don't need the source code.
2. I am a developer of Antha: I need the source code.
3. I want to build a docker image which contains Antha.

Antha is built using [go mod](https://github.com/golang/go/wiki/Modules) to manage dependencies.
You do not need to worry about `GOPATH` environment variables, and you can keep the source to Antha anywhere, if you need it.

### Pre-requisites

 * Ensure that your installed version of Go is at least 1.12.4.
 * Ensure that git has been installed.
 * Ensure that your git installation is configured properly to allow access to Synthace repositories.

#### Installing Go

Follow [the official instructions](https://golang.org/doc/install).
Note that an alternative on Mac OS is to install and manage go using *home brew* - however this will not play nicely with a manual install. (i.e. Pick one installation method, don't mix and match.)

### Install Antha without source code

This allows installation of the Antha commands without needing to explicitly download the code.

1. Create a directory for installation. This may be wherever convenient.
  a. Apple/*nix : `mkdir antha`
  b. Windows : Via UI, or `mkdir antha`
  NOTE: All following steps should take place within this directory. (`cd antha` to enter this directory now.)
2. From within this directory, initialise the project:
  `go mod init Antha`
3. Add a dependency to the project:
  `go mod edit "-require=github.com/antha-lang/antha@feature/future_sanity"`
  NOTE: **feature/future_sanity** may be edited to refer to any branch, or commit SHA if a particular version of Antha is desired.
4. Install the Antha commands:
  `go install github.com/antha-lang/antha/cmd/...`

This has now installed the [antha commands](#antha-commands) ready for use.

#### Upgrading (Or changing version)

All operations should be run from the install directory created when installing.

1. Tidy the current installation : `go mod tidy`
2. Get the required version of antha-lang using `go get`
    e.g. Latest master : `go get github.com/antha-lang/antha@master`
    e.g. Latest version of a branch : `go get github.com/antha-lang/antha@my/branch-name` (for branch _my/branch-name_)
    e.g. A particular commit : `go get github.com/antha-lang/antha@f5472259fc2eec9f7443f4d5f56c6739a1d9e4db`  (Where `f5472259fc2eec9f7443f4d5f56c6739a1d9e4db` is the SHA for a particular commit)
3. Re-install the tool set:
  `go install github.com/antha-lang/antha/cmd/...`

#### Uninstalling

Installing Antha adds various commands to your system. You list them by running the following command from the install directory created when installing:

  1. `go list -f '{{ .Target }}'  github.com/antha-lang/antha/cmd/...`
  2. Each line from the above command gives the location of a command installed by Antha on your system. You can delete these if they exist.

### Install Antha with source code

This is the recommended installation if source code is required, for development of Antha. Due to the use of `go mod`, installation must *not* be under `$GOPATH`

The following projects need to be cloned into the same directory:

  - `git clone -b feature/future_sanity git@github.com:antha-lang/antha.git`
  - `git clone -b feature/future_sanity git@github.com:Synthace/antha-runner.git`
  - `git clone -b feature/future_sanity git@github.com:Synthace/instruction-plugins.git`

(Note, that local dependencies may be observed in `antha/go.mod` - those redirected to local copies should be cloned directly. At current time this is `antha-runner` and `instruction-plugins`)

Once done, usual Go compilation is possible:

  - Install local versions of executables:
    * Within the required command directory, e.g. `antha/cmd/composer`, build the command.
    * `go build`
    * Executables are now available in the local directory (e.g. `antha/cmd/composer/composer`)
  - Install all commands into the environment:
    * Within `antha`
    * `go install ./cmd/...`
    * Installs all Antha commands into the path.

## Antha Commands

Installing Antha adds a number of commands to your system. Source code for each may be found within the `cmd/` directory, and associated documentation in a `doc.go` file. This is in standard Go documentation format.

`go doc` files may be viewed from the command line using the command `go doc -all package/name` (e.g. `go doc -all github.com/antha-lang/antha/cmd/composer`)

Besides viewing help for the Antha commands, information on antha-lang code may also be inspected.
e.g. for information on the representation of workflows: `go doc -all github.com/antha-lang/antha/workflow`

### General Usage

The following commands are installed:

  * [composer](#using-the-composer-command) for composing and running workflows.
  * [elements](#using-the-elements-command) for managing elements.
  * [migrate](#migrating-old-workflows) for migrating old workflow data into current format.

Antha commands generally allow for workflow files to be composed. This is useful during development, as it allows different parts of a workflow to be shared between different workflows. For example:

 * [Device configuration](cmd/composer/gilsonOnly.json.sample)
 * [Repository information](cmd/composer/repositories.json.sample)

#### Use of repository json

Workflow json now contains detailed repository information to be used when accessing elements. This detail allows for reproducible workflows (the same version of elements will always be pulled), for targeting specific versions of a file, or even mixing repositories from multiple locations.

See `go doc` documentation with `go doc -all github.com/antha-lang/antha/workflow.Repository`

See [example here](cmd/composer/repositories.json.sample).

You will need to copy the [cmd/composer/repositories.json.sample] file to `repositories.json` and edit it, ensuring that the `Directory` field matches your clone of the elements repository, and the `Branch` field refers to the branch (element set) with which you are working. You will then use this `repositories.json` file as an argument to many of the Antha commands.

#### Use of device configuration

Configuration of devices is now managed at a device level, not at a workflow level. This allows for variation across multiple devices.

See [example here](cmd/composer/gilsonOnly.json.sample).

More detailed information may be found by either:
  * `go doc comments`
    * `go doc github.com/antha-lang/antha/workflow.Configuration` for general configuration.
    * Other recognized configuration may further be explored with `go doc` by following the relevant types, for example
      `go doc github.com/antha-lang/antha/workflow.GilsonPipetMaxInstanceConfig` for the configuration specific to
      gilson instances, or `go doc github.com/antha-lang/antha/workflow.CommonMixerInstanceConfig` for generic device
      configuration.

  * The [JSON schema for the workflow](workflow/schemas/workflow.schema.json)

#### Using the `composer` command

See `go doc github.com/antha-lang/antha/cmd/composer` for detailed documentation.

The `composer` command composes together workflows, and optionally executes them. Flags are available to control exact behaviour (for example skip executing a workflow, retain transpiled element sources,...). At a very high level, the composer will:

  * Take a set of workflows and combine them.
  * Extract element sources from repositories as necessary.
  * Set up required devices.
  * Generate executable code.
  * Execute the workflow.

The input is a workflow, often split into multiple json for [device](#use-of-device-configuration) and [repository](#use-of-repository-json) information.

Note that there are currently multiple options for communicating with instruction plugins:

  * Use `-linkedDrivers=true` (the default). This requires that the `github.com/Synthace/instruction-plugins` repository is available on your system.
  * Use `-linkedDrivers=false`. This relies on the `Connection` field of the device configuration.
      * The options for `Connection` are:
        * A local executable which you have prepared, preceded by `file://` - e.g. `file:///Users/someone/go/src/github.com/Synthace/instruction-plugins/PipetMax/PipetMax`
        * A go directory preceded by `go://` containing source code which will be built by the composer - e.g. `/Users/someone/go/src/github.com/Synthace/instruction-plugins/PipetMax`
        * A port over which an existing instruction plugin is running - e.g. `localhost:50051`

#### Using the `elements` command

See `go doc github.com/antha-lang/antha/cmd/elements` for detailed documentation.

The `elements` command allows for inspection and manipulation of elements within a repository. For example:

  * _List_ elements available in a repository
  * _Describe_ elements available in a repository (i.e. usage information about the element)
  * _Make_ a workflow using elements. This is particularly useful for testing - a workflow may be constructed instantiating given elements.
  * [Test elements in a repository](#running-element-tests).

### Migrating Old Workflows

Workflows in previous formats must be upgraded before they can be used with the Antha commands. The `migrate` command exists to help with this task. One of the more significant changes to the workflow format is that file content is no longer part of the JSON. Instead, it is stored externally to the JSON, and so a complete workflow is now a directory rather than an individual file.

#### Using migrate

The safest migration method is to use the tool `migrate`. For detailed documentation see `go doc github.com/antha-lang/antha/cmd/migrate`.
This will migrate a workflow:

  * A json snippet including repository information must be supplied, as this information is not available in older files.
    * It may be desirable to delete the repository information in the migrated file, so that it may be combined easily with multiple repository fragments later.

  * Further snippets may be supplied containing device information. Configuration will then be applied to the supplied devices.
    * Alternatively, a new device may be created using the command line flag `-gilson-device`. This newly created device will receive migrated device configuration.
    * There is no option to automatically create non-gilson devices, as these were not present in previous workflow json formats.

  * The migration tool will move embedded data in external files. Antha does not support embedding data directly in workflow json files.

#### Running Antha unit tests

This is the standard method of running Go unit tests.

* For installations including source code:
    * `go test ./...` from within the installed `antha` directory.

* For [sourceless installations](#install-antha-without-source-code):
    * `go test github.com/antha-lang/antha/...`

#### Running element tests

Given an element repository, this may be tested using the standard Go testing framework:
  * Test that all elements may be compiled. (Temporary workflows including all elements are generated and compiled.)
  * Find all standard go tests, and run them.
  * Find any worflows within the repository, and run them.
  * For any test workflows containing comparison data, check that the restult of running the workflow matches the comparison data.

Element tests are called via running `go test` on the `elements` command, supplying a repository description to run tests on. (See also `go doc github.com/antha-lang/antha/cmd/elements`)
For installations including source code, the following may be run from the `antha/cmd/elements` directory:

    `go test -v -args -keep -outdir=/tmp/foobar ../composer/repositories.json`

Which indicates test output to be written to `/tmp/foobar` and repositories to be tested are specified in `../composer/repositories.json`. The `outdir` must be empty or non-existent and a lot of output will be produced. So it may be convenient to include commands to clear the output directory and save all output to a file:

    `rm -rf /tmp/foobar && go test -v -args -keep -outdir=/tmp/foobar ../composer/repositories.json 2>&1 | tee /tmp/log`

If the installation is [sourceless](#install-antha-without-source-code) then the elements package should be specified on the command line:

    `rm -rf /tmp/foobar && go test github.com/antha-lang/antha/cmd/elements -v -args -keep -outdir=/tmp/foobar ../composer/repositories.json 2>&1 | tee /tmp/log`

## Updating dependencies

Go will suggest dependency updates [as described here](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies) - for example use `go get -u` to get the latest version of direct and indirect dependencies.

## Editing Workflows

In order to support editing workflows, support is given in the form of a [json schema definition here](workflow/schemas/workflow.schema.json). This is used for validation by Antha commands, and may be used to assist in creation and editing of workflows - use of json schemas depends on your editor choice.

## Local edits to `mod.go`

Go modules ensure that builds are reproducible, however it may be desirable to edit local versions of certain includes.
Certain modules (i.e. _antha-runner_ and _instruction-plugins_) are already managed locally, however it may be useful to edit further module sources - in this case the `go.mod` file may be edited to reference alternative copies.

Further, there are tools e.g. [gohack](https://github.com/rogpeppe/gohack) which may be of use in helping to manage local dependency edits.

## Building a Docker image containing Antha

This is slightly more advanced, but may be convenient for controlled environments. (It is extensively used by Antha backend, and is an actively maintained method.)

Full instructions may be found in [this readme](README_Dockerfile.md)
