// The Migrate Command
//
// Example:
//   migrate -from=path/to/old.json -format=json myRepositories.json
//
// The migrate command migrates workflows from outdated historical
// schema versions to the current schema version. Currently, it will
// accept as inputs:
//
//  version=1.2 workflow files in JSON format (the old "bundle" format)
//  version=1.2 SimulateRequest protobuf objects
//
// and will produce SchemaVersion=2.0 workflows.
//
// The version=1.2 format does not contain any versioning information
// for elements. Thus when migrating, as a minimum, you are required
// to provide Repository information in the SchemaVersion=2.0 format
// such that each element type can be located within one of the
// repositories specified.
//
// The following flags are available:
//
//  -from=path/to/file
//    The workflow to migrate. Use - to read from stdin.
//
//  -outdir=path/to/directory
//    A directory to write the results to. In SchemaVersion=2.0, the
//    workflow.json itself cannot contain file content. Thus a
//    directory must be provided so that file content that was baked
//    into the old workflow can be extracted and written out. The new
//    workflow is written to directory/workflow/workflow.json, and any
//    file contents are written to directory/data/. The new workflow
//    will contain references to any extracted files, and the
//    directory layout matches the requirements of the composer -indir
//    flag. If not provided, a fresh temporary directory is used.
//
//  -validate (boolean, default true)
//    Whether or not to attempt to validate the migrated workflow. In
//    some cases it may be necessary to disable validation if it is
//    known that the generated workflow is incomplete (e.g. you know
//    you're only producing a workflow snippet).
//
//  -format=json (enum, default is json, other legal value is protobuf)
//    Indicate the format of the input workflow to migrate. If json,
//    then the input is parsed as a version=1.2 JSON workflow file. If
//    protobuf, then the input is parsed as a version=1.2
//    SimulateRequest protobuf object.
//
//  -gilson-device=myFirstGilson (optional, json only)
//    In version=1.2 workflows in JSON format, the only supported
//    liquid handler device is the Gilson PipetMax, and only one such
//    device is supported per workflow, and the device is unnamed. To
//    migrate those configuration values to corresponding entries in
//    the current SchemaVersion=2.0 workflow, a device name must be
//    provided. If no such device name is provided (which is the
//    default), then Gilson PipetMax configuration parameters are not
//    migrated.
//
//    For protobuf inputs, the first mixer device only is migrated,
//    and all non-mixer devices are migrated. This matches the
//    historical behaviour of antha.
//
// Additionally, as normal, workflow snippets may be provided as
// additional arguments to the command.
//
// Log messages are produced on stderr.
package main
