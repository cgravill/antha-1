// The Elements Command
//
// The elements command is a command to help work with elements and
// workflows. Various subcommands are available, and just like the
// composer command (github.com/antha-lang/antha/cmd/composer), they
// accept workflow snippets as inputs.
//
// Additionally, this package facilitates the testing of elements
// through the usual `go test` mechanism, documented below.
//
// In all cases, log messages are produced on stderr.
//
// List
//
// The list subcommand lists elements in a structured way. This makes
// its output useful for consumption by other tools. The output (on
// stdout) has one line per element type, with 3 fields,
// tab-separated: the element type name, the element type path
// (i.e. path within the repository to the element type's directory),
// and the repository name.
//
// Workflow snippets may be provided both on the command line, or the
// -indir flag may be used to provide a path to a directory directly
// containing workflow snippet json files.
//
// Additionally, a -regex flag may be used, the value of which is a
// regular expression (https://github.com/google/re2/wiki/Syntax), and
// is matched against the element type's path. Only matching element
// types are output. The default is to match all element types.
//
// Example:
//    elements list -regex=Aliquot path/to/repositories.json
//
// Describe
//
// The describe subcommand is similar to the list subcommand, but for
// each matching element type, it outputs all the known metadata and
// documentation for the element type.
//
// It accepts -indir and -regex flags exactly as for the list
// subcommand, and workflow snippets can be provided in exactly the
// same way. For every found element type that matches the regex
// (default is to match all element types), the following information
// is presented:
//
//  - Element type name
//  - Repository name
//  - Element path
//  - Description (documentation of the element type)
//  - Ports
//    For every field within the element type's Inputs, Parameters,
//    Outputs, and Data, the field name, any defaults if known (taken
//    from the corresponding metadata.json if found), and the
//    documentation of the field.
//
// The information is output via stdout, and is appropriately tab
// indented.
//
// Example:
//    elements describe -regex=AddFluorescenceTimeCourse path/to/repositories.json
//
// Make Workflow
//
// The makeWorkflow subcommand allows for workflows to be generated
// containing instances of particular element types. This is very
// useful when testing or developing elements in order to check that
// antha element code transpiles and compiles correctly.
//
// It accepts -indir and -regex flags exactly as for the list
// subcommand, and workflow snippets can be provided in exactly the
// same way.
//
// Additionally, a -to flag may be used to specify a file to write
// output to. The default is to output to stdout.
//
// For every found element type for which the element type's path
// matches the regex, the generated workflow will contain a
// corresponding Element Type entry, and a corresponding Element
// Instance entry. No parameters are set on the element instance, and
// no connections are made between any elements.
//
// Example:
//
// To generate a workflow containing all qpcr elements and check those
// elements transpile and compile correct, a command line such as the
// following could be used:
//
//  elements makeWorkflow -regex=QPCR path/to/repositories.json | composer -run=false -keep -
//
// Testing
//
// Whilst not a subcommand, testing of elements is supported by this
// package.
//
// Testing accepts the -indir and -regex flags exactly as for the list
// subcommand, and workflow snippets can be provided in exactly the
// same way. The -regex flag is only used when selecting test
// workflows to run.
//
// Testing also accepts an -outdir flag which can be used to provide a
// path to a directory under which all source checkouts, transpilation
// and compilation will occur. If not provided, the default is to use
// a fresh temporary directory.
//
// There are 2 distinct stages of testing:
//
//  1. TestElements/CompileAndTest
//   This internally uses makeWorkflow to generate a workflow
//   containing an instance of every known element type. This is then
//   transpiled and compiled in the normal way (the generated workflow
//   is not executed).
//
//   Next we attempt to run every go test within the checked-out and
//   transpiled source directories. Thus normal go tests within the
//   elements repositories are run.
//
//  2. TestElements/Workflows
//   For the purposes of this documentation, a bundle is a workflow
//   which also contains expected outputs of the planner.
//
//   For every bundle which we find within the source repositories, we
//   compose, validate, transpile, compile and execute the
//   workflow. The output of the planner is compared to the expected
//   outputs provided by the bundle file, and thus the success of the
//   test is determined.
//
//   For this test, there is a -regex flag, which if provided is
//   matched against the path of the bundle.
//
// Examples:
//
//  Run all the tests described above:
//
//    go test -v -args -outdir /tmp/myTests repositories.json
//
//  Run only the Workflow test, and within that, only those workflows
//  who's path contains AddFluorescenceTimeCourse:
//
//    go test -v -run=Elements/Workflows -args -regex=AddFluorescenceTimeCourse -outdir=/tmp/myTests
//
// The -args flag is necessary to separate args to the go test system
// from args to our tests.
package main
