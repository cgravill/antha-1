# Antha Quickstart

This quick start guide describes a basic *local* Antha tools installation and a way of using it. 

A more complete guide can be found in the [main readme](README.md), including alternative options for
installation and usage. It walks through examples for the key features of the Antha tools. 

This guide will assume that:
  * You're familiar with the command line of your operating system. (Mac OS/Windows/Linux/Other)
  * You have `git` installed, and know basic usage.
  * You have an up to data `go` installation. (See [main readme](README.md) for more details.)

It's recommended that this guide is worked through in order, as some parts build on previous work (e.g. you need to install Antha, before you can run a workflow.) 
The sections are:

  * [I want to install Antha](#I-want-to-install-antha)
  * [I want to run a workflow](#I-want-to-run-a-workflow)
  * [I want to test a repository](#I-want-to-test-a-repository)
  * [I want to migrate an old workflow](#I-want-to-migrate-an-old-workflow)
  * [What next?](#What-next)

## I want to install Antha

### 1 Create a working directory

Let's get started. I'm picking my home directory, and creating a new folder called `antha`:

```bash
cd ~
mkdir antha
cd antha
```

You can pick any directory you like, **BUT** it should not be withing your `GOPATH` (If you're not sure what your go path is, type `go env` at a prompt, and look for the line `GOPATH=` to find your GOPATH.)

**NOTE** I'll often refer to this directory we've just created (`antha`) as the _Antha working directory_ or _working directory_. For the quickstart examples, we'll generally do all of our work in here, and instructions at the command line should usually be typed in while in this directory.

### 2 Clone required repositories

We've got a nice clean working area, lets get the code that we need!

```bash
git clone https://github.com/antha-lang/antha.git
git clone https://github.com/Synthace/antha-runner.git
git clone https://github.com/Synthace/instruction-plugins.git
```

Lets also get an elements repository to work with, I'm picking elements-westeros on the development environment:

```bash
git clone https://repos.antha.com/antha-ninja/elements-westeros.git
```

This should have got copies of all the latest antha folders locally. I should see the following folders:

```
antha
antha-runner
instruction-plugins
elements-westeros
```

#### Troubleshooting

 * I can't clone the repositories!
   * It's likely to be a permissions problem. Check that you can access these repositories in a web browser:
   * https://github.com/antha-lang/antha
   * https://github.com/Synthace/antha-runner.git
   * https://github.com/Synthace/instruction-plugins.git
   * https://repos.antha.com/antha-ninja/elements-westeros
   * If you're not able to access these, then you'll need to request access from someone who does.
   * If these are all fine, then it's likely that `git` is not configured correctly on your machine - again
     best to discuss with someone with a working environment.

### 3 Housekeeping

That's great - we've got copies of the files that we need. **BUT** we are now also responsible for them.

  * If we want to use different versions of antha, we'll need use `git` to get the versions that we want and then [reinstall the tools](#4-Install-the-tools).
  * If we want to edit an antha build, we'll need to make our edits, and then [reinstall the tools](#4-Install-the-tools).
  * If we want to use different elements repositories, we'll need to use `git` to get the versions that we want.
  * If we want to make local edits, we'll need to make edits in that folder.

**NOTE**: Unlike older versions of Antha, there is no requirement to _build_ elements. Antha tools will compile the 
elements for you at the appropriate time.

At the current time, we'll want to shift to the `feature/future_sanity` branch, so lets do it. We'll need to switch all of our repositories to this branch. Assuming you're in [your working directory](#1-Create-a-working-directory), enter the following:

```bash
cd antha
git checkout feature/future_sanity
git pull
cd ../antha-runner
git checkout feature/future_sanity
git pull
cd ../instruction-plugins
git checkout feature/future_sanity
git pull
cd ../elements-westeros
git checkout feature/future_sanity
git pull
cd ..
```

### 4 Install the tools

We've got the code, lets build and install the tools. Assuming you're in [your working directory](#1-Create-a-working-directory), enter the following:

```bash
cd antha
go install ./cmd/...
cd ..
```

This may take some time the first time (while `go` downloads the required packages), but should be quick on subsequent attempts.

**NOTE**: If you make any changes in `antha`, `antha-runner` or `instruction-plugins` either local edits or via `git`, you will need to re-run these commands. (This will rebuild the tools, taking in any changes.)
**NOTE**: It is always safe to run these commands! If ever in doubt about whether you're up to date, run them again.

#### What has this got me?

Good question. You now have three tools installed in your system. 

  * **composer**
    * Used to run workflows.
  * **elements**
    * Used to investigate and test element repositories.
  * **migrate**
    * Used to migrate old versions of workflows.

As we've installed them, they're always available from the command line. Let's try. Type:

```bash
composer
```

You should see something like:

```
ts=2019-05-16T11:21:23.109475Z fatal="No workflow sources provided."
```

True, but not so helpful. Let's look at the help:

```bash
composer --help
```

That's a bit more useful, you should see some information about the syntax for the command.
We can also check the go documentation for the command. Run the following from your [working directory](#1-Create-a-working-directory)

```bash
go doc ./antha/cmd/composer
```

Ok, that's got more detail and usage. For the installed commands you can always use the _command help_ and _go documentation_ to check the up to date documentation for a command. For example:

```bash
elements --help
```

or 

```bash
go doc ./antha/cmd/elements
```

That completes installation of our tools, we can move on to using them.


### 5 Recap

During installation we:

* [Created a working directory](#1-Create-a-working-directory)
* [Got a copy of the Antha Code](#2-Clone-required-repositories)
* [Got a copy of an elements repository](#2-Clone-required-repositories)
* [Checked we had the right versions of our code and elements](#3-Housekeeping)
* [Installed the tools](#4-Install-the-tools)
* [Checked that we can access help and documentation for the tools](#What-has-this-got-me)

We're all set to do some work

## I want to run a workflow

Let's run a pre-existing workflow. We'll take one of the test workflows from the `elements-westeros` repository that we have. The workflow that we're going to run is:

```
elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow
```

If you look in this file, you'll see the following files:

 * `data/1` - This is a data file for the workflow
 * `workflow/workflow.json` - This is the workflow definition, it specifies the elements used, how they're connected, parameter values and device configuration.

### 1 Run the workflow

We'll run using the `composer` commmand, but we'll need to do some setup first. In the new antha world, a workflow is
completely self contained - it contains all the information required to run the workflow in a precise way. See the [main README](README.md) for more details on contents, but briefly it consists of a number of data files (in a data directory) and `json` file detailing:

  * The workflow design - elements, how they're connected, what parameters they have.
  * General configuration.
  * Device configuration. 
  * Where to get elements from.

For the test workflow that we're looking at, we have data file (`data/1`) and a workflow json (`workflow/workflow.json`) *but* the `workflow.json` is incomplete - it doesn't contain information on where to get elements from!

This makes sense, users will often want to run a single workflow against different element sets, particularly during development or testing. We could fix this by editing `workflow.json` to put in the information, but `composer` gives us a cleaner option - we can supply an extra `json` containing just the missing information and `composer` will merge this information in before running.

#### Create a local repository json

Let's create our `json` with repository information. Create a file `local-repository.json` in your working directory, and edit so that it reads as follows:

```json
{
    "SchemaVersion": "2.0",
    "Repositories": {
        "repos.antha.com/antha-ninja/elements-westeros": {
            "Directory": "/Users/tutor/antha/elements-westeros"
        }
    }
}
```

**IMPORTANT** You will need to change the **Directory** (here `/Users/tutor/antha/elements-westeros`) to wherever your copy of `elements-westeros` is. You cannot use the `~` notation for a home directory here.

We can now use this to run the workflow, using this new file to indicate where our elements are installed. Run the following from your [working directory](#1-Create-a-working-directory):

```bash
composer -linkedDrivers -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  ./local-repository.json
```

That's it - it should run the workflow. You'll see lots of output to the command line, ending with something like:

```
...
requested 19.850000000000307 ul from well "B1@auto_input_plate_1_mammal" which only contains 19.350000000000307 ul working volume and 69.3500000000003 ul total volume
ts=2019-05-16T12:09:35.959003Z progress=complete
```

On running the first time it may need to download dependencies so make take longer than usual.

Lets take the command apart a little to understand it:

```bash
composer 
[1]  -linkedDrivers 
[2]  -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  
[3]  ./local-repository.json
```

 1. `-linkedDrivers` - this instructs antha to build the drivers from `instruction-plugins`. This is the recommended method for local development. If this is missing, you may see error messages about being unable to communicate with 
 device drivers.
 2. `-indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow`. This tells `composer` to take the contents of this folder as a workflow. This ensures that we include data files. **All** `.json` files within this directory will be considered part of the workflow, and be merged together. (i.e. Multiple incompatible workflow .json, or non-antha .json files will cause problems.)
 3. `./local-repository.json`. Any files listed at the end will be merged into the workflow. Here we're merging in the information about where our elements repository is located.

### 2 Make the run easier to work with 

We've successfully run a workflow - but where's the output? What happened? How can I debug if things go wrong? There are a number of options to make running the workflow easier to handle.

#### Keep the workflow

Normally `composer` discards successful workflows! Often we want to _keep_ the results. Fortunately the `-keep` flag does exactly that. Run:

```bash
composer -linkedDrivers -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow -keep ./local-repository.json
```

Notice that we've added `-keep`. It runs the same, but where *is* the output?
If you scroll back, you'll somewhere near the top:

```
ts=2019-05-16T12:25:53.972528Z outdir=/var/folders/mj/96w6t_gx37sgc813j3ss4g5r0000gn/T/antha-composer423580724
```

Or maybe not, if it's scrolled too far. 
`composer` has automatically created an output directory, somewhere that the OS has decided is appropriate. It would be useful if we could tell it where to send output - let's do that!

#### Specify an output directory

`-outdir` option gives us exactly what we need. Lets say that we want to write all our output to a folder named `output`.
Our new command line looks like:

```bash
composer -linkedDrivers -outdir=./output -keep -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  ./local-repository.json
```


Note the addition of `-outdir=./output`. (_Note also_ that it must come **before** `./local-repository.json` - otherwise `composer` will assume the extra text is a workflow json to be merged in.)

We should see that an `output` directory has been created. Lets check the contents:

```bash
ls output
```

We should see:
```bin		logs.txt	simulation	src		workflow```

That all looks useful! We'll come to what it means in [a later section](#What-is-in-this-output), but for now let's finish tidying how we run this.

#### Other tidy ups

I'm excited! Let's run the workflow again:

```bash
composer -linkedDrivers -outdir=./output -keep -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  ./local-repository.json
``` 

Oh dear! Assuming you're following along you'll see something like:

```
ts=2019-05-16T12:37:41.797346Z fatal="Provided outdir './output' must be empty (or not exist)"
```

`composer` is cautious - it will refuse to overwrite any information that already exists, in case you still need it. As a result you'll have to delete it yourself. Let's try again:

```
rm -rf ./output
composer -linkedDrivers -outdir=./output -keep -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  ./local-repository.json
``` 

That looks better.
If we're going to be running this a lot, you might want to chain these together in one line:

```
rm -rf ./output && composer -linkedDrivers -outdir=./output -keep -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  ./local-repository.json
```

We can now repeatedly run a workflow, but there's an awful lot of output in the console. Let's pipe it somewhere so that I can look at it later. I've chosen `run-output.txt`.

```
rm -rf ./output && composer -linkedDrivers -outdir=./output -keep -indir=./elements-westeros/DemoWorkflows/DOE/SimpleExample/testdata/defaultDOEExample_workflow  ./local-repository.json > run-output.txt 2>&1
```

(Note the addition of `> run-output.txt 2>&1` at the **END** of the command. This sends both normal and error messages to the specified file.)

We can check the output from the command at our convenience:

```
less run-output.txt
```

Or whatever editor you prefer.

### What is in this output?

We now have lots of output in `output`. What does it mean?

 * `logs.txt` - exactly what it sounds like, the logs from the run. This should usually be the first thing to check.
 * `bin` - This contains the executable generated from the workflow. Probably not of immediate interest.
 * `simulation` - This contains the output of the simulation.
 * `src` - This contains copies of all the input files, including JSON, data files and copies of relevant elements. Can bbe worth checking to ensure that the expected files were used. 
 * `workflow` - This contains the _compiled_ version of the workflow in `main.go` - it can be useful to check this to ensure that the compiled code is as expected.	

### Recap

This section has walked through running a workflow using `antha`. It has covered:

  * How to pick a workflow to run
  * How to add missing information to a workflow
  * How to run the workflow using `composer`
  * Detail about the behaviour of `composer` and how to run it in a useful way.
  * What the output means.

A few things that you could do now:

  * Edit elements, then re-run the workflow with those new elements.
  * Edit a workflow, then re-run it to see the effect.
  * Make edits to antha itself, rebuild the tools, and re-run a workflow.

## I want to test a repository

This is all well and good, but what about testing a large collection of workflows? For example you edit some elements and
want to check that you haven't broken anyone else's workflows?

Luckily this also is included in the antha tool set - it's possible to test a large number of workflows, and more, in a single command. This functionality utilises the existing [go testing framework](https://golang.org/pkg/testing/), so besides running workflow tests you can write normal go tests.

### Test a repository

Let's test the repository that we have. From your working directory, type the following:

```bash
rm -rf test-output
go test ./antha/cmd/elements -v -args -keep -outdir=`pwd`/test-output `pwd`/local-repository.json
```

This may take some time to run, there are over 150 tests in the standard Antha repository.
The test framework is doing the following for you:
  * Locating all elements in the repository, and check that they are valid code.
  * Locating all _go_ tests in the repository (i.e. element unit tests) and running them.
  * Locating all workflows in the repository, run them, and check outputs if output data has been supplied.

Let's disect the commands to understand what's going on:

  * `rm -rf test-output` - this is to clean out any already existing test output. (The tests will fail if there is pre-existing test output.)
  * `go test ./antha/cmd/elements` - runs the golang tests in `elements` (This elements directory contains the test logic that we need.)
  * `-v` - turn on verbose output.
  * `-args` - specifies that all following args are Antha specific controls.
  * `-keep` - keeps all output (rather than deleting successful runs).
  * ``-outdir=`pwd`/test-output`` - specifies that we'll write all results to `test-output` directory. (This will be created if it doesn't already exist.) **NOTE** the `` `pwd` `` has been used to add the current directory to the output path. A full path should be supplied here, due to `go test` changing directories. We could instead use any absolute path for our output directory.
  * `` `pwd`/local-repository.json `` - specifies the json file specifying a repository to test. **NOTE** that `` `pwd` `` has been used here, to add the full directory information for `local-repository.json`. Because the go tests change directory, the full path to the json file should be supplied. (Generally, instead of `` `pwd` `` we could use `/Users/tutor/antha/local-repository.json`, or your equivalent.)

Similarly to the way we [improved our command line for composer](#other-tidy-ups) we may want to do similar here:

```bash
rm -rf test-output && go test ./antha/cmd/elements -v -args -keep -outdir=`pwd`/test-output `pwd`/local-repository.json > test-output.txt 2>&1
``` 

Where we have:
  * Combined our statements into a single line.
  * Piped the output to `test-output.txt` to make it easier to inspect after.

### What is the test output?

There is a *lot* of information generated by test runs. We'll give an overview of the key areas:

  * `test-output.txt` (if run as above) - this is the output from the test run. This is usually the first thing to look at.
    * The end will indicate a `PASS` or `FAIL`
    * Failing tests may easily be found by searching for `FAIL:`. Following this you should see information about the name of the test, and some details of failure.

The remaining data is found in the test output location - when run as above, this is `./test-output`
  * `/testWorkflows` - this folder contains the results of running the discovered workflows.
    * `logs.txt` - this is the log from running all the workflows. This is usually the first thing to check.
    * `/simulation` - this is usually the second thing to look at, once you know which tests are failing. It contains the individual results from each workflow that was run - the output is [as previously discussed here](#What-is-in-this-output).  
    * `/src` - copies of all the source files used in the run.
    * `/workflow` - the generated workflows.
    * `/bin` - the compiled versions of drivers if used.


  * `/compileAllElements` - this folder contains the results of collecting all elements in the repository and attempting to compile.
      * `logs.txt` - logging from compiling all elements. This is usually the first thing to check.
      * `/src` - this is a copy of the repository source under test.
      * `/workflow` - this is the generated workflow containing all found elements.
      * `/bin` - this is the compiled version of the workflow.
  

## I want to migrate an old workflow

There are several cases where you might have an old Antha workflow, but want to upgrade it by hand. Fortunately there's a tool to  do it for you. This is the `migrate` tool installed with Antha.

  * The oldest supported version of antha workflows (sometimes referred to as _bundle files_) is **1.2**
  * The _current_ Antha workflow version is **2.0**

### Getting an example of an old workflow

Lets get an example workflow to migrate. Note that this is only for the migration example and is not generally required.

We'll take a historical copy of the antha elements, which is using 1.2 format. In your working directory, run the following:

```bash
git clone https://repos.antha.com/antha-ninja/elements-westeros.git old-elements-westeros
cd old-elements-westeros
git reset --hard 051a87ef3
cd ..
```

You will now have a historical copy of Antha elements in `old-elements-westeros`. We'll migrate our workflows from here - feel free to delete `old-elements-westeros` once finished with examples.

### Basic migration

The `migrate` tool will do most of the work for you, but the **1.2** is missing two pieces of information:

  1. Information regarding how to reference elements within a repository. (It just has a list of element names.)
     To fix this, we will provide a json file containing repository information.
  2. A named gilson device.
     To fix this, we will provide a device name to the `migrate` command. *NOTE* As we are only simulating on the command line, not running on a real machine, we may pick any name for our device.

If you've been following along, you should have a file `local-repository.json` from the step [Create a local repository json](#Create-a-local-repository-json). If not, or it has been edited, then recreate it from [that step](#Create-a-local-repository-json).

That's everything we need! From your working directory, run:

```bash
migrate -from=old-elements-westeros/DemoWorkflows/qPCR/Large_Sample_Set_QPCR_DemoWorkflow.jso -outdir=migrated -gilson-device=gil  ./local-repository.json
```

We have specified:
  * `-from=old-elements-westeros/DemoWorkflows/qPCR/Large_Sample_Set_QPCR_DemoWorkflow.json` - an old workflow json to migrate.
  * `-outdir=migrated` - a target directory for the migrated workflow. (This will be created if it doesn't exist.)
  * `-gilson-device=gil` - a name for a new gilson device to create. This will receive migrated configuration from the old workflow.
  * `./local-repository.json` - a workflow file in the **new** format, to supply extra information, in this case repository data.

 a local output directory of `migrated`. Let's take a look inside - you should see files:

```
migrates/workflow/workflow.json
data/1
```

It is important to note that while the original workflow was a single json file, the output is a **directory** combining `workflow.json` and data.
This may now be run as in [the description here](#I-want-to-run-a-workflow). For example, type the following in your working directory to run:

```bash
rm -rf ./output && composer -linkedDrivers -outdir=./output -keep -indir=./migrated
```

While that completes the migration, we make some further observations.

#### Observation : Extraction of embedded data

It's important to note that the migrate command has not only migrated the workflow, but also moved data which was embedded in the workflow (as `QPCRDesignFile`) to the external file `data/1`. 

Notice that the old `Large_Sample_Set_QPCR_DemoWorkflow.json` contained a section:

```json
"QPCRDesignFile": {
        "bytes": {
          "bytes": "UmVhY3Rpb..."},
        "name": "Large_QPCR_Design_File_.csv"
      }
```

while the new `workflow.json` contains a reference to the external file:

```json
"QPCRDesignFile": {
    "Path": "1",
    "IsOutput": false,
    "Name": "Large_QPCR_Design_File_.csv"
}
```

At the same time, while the old embedded version of the file was encoded as `UmVhY3Rpb...`, the external file `data/1` is in the original, readable format:

```
ReactionName,Template,Primer1,Primer2,Dilution,Number of dilution replicates,Working solutions,Dilution Factor,Diluent Name,Sample Type,Probe,Replicates,Mastermix
Reaction1,IL4,IL4_FWD,IL4_RVS,3,1,2,10,DiluentA,Sample,SYBR,3,Mix A
...
```

This, and the fact that the `workflow.json` is kept small, leads to more managable workflows.

#### Observation : Removing repository information

`migrate` will always insert repository information into the migrated workflow, but you may wish to remove it. For example, this may be a test workflow, and you want the testing framework to select the appropriate element repository as appropriate. Alternatively you may be sharing a workflow with others, or working with multiple repositories - in each case it is likely easier to keep repository information out of `workflow.json` and only supply when running a workflow.

The most direct method is simply remove the `Repositories` element from the workflow json. It will look something like:

```json
        "Repositories": {
                "repos.antha.com/antha-ninja/elements-westeros": {
                        "Directory": "/Users/matthewgregg/antha/elements-westeros"
                }
        },
```

(Advanced users may prefer to use the [jq tool](https://stedolan.github.io/jq/) to manipulate json to achieve the same result.)

Once this has been removed, we then need to supply the missing information whenever running the workflow, for example:

```bash
rm -rf ./output && composer -linkedDrivers -outdir=./output -keep -indir=./migrated  ./local-repository.json
```

Note the addition of `./local-repository.json` at the end of this, to add back the repository information.

#### Observation : Other migration options

While the basic migration above will be sufficient for most cases, there are some other options:

 * Disable validation with `-validate=false` . This will not check whether the generated workflow is valid. This may be of use if you know that the original workflow has problems, or missing information. If validation is disabled, then the generated output is not necessarily runnable without manual edits. 

## What next? 