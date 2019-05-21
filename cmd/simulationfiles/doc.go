// The SimulationFiles Command
//
// Example:
//   simulationfiles -outdir=/tmp/extracted/ -indir=antha/simulation/workflow/
//
// The simulationfiles command exists to extract files that were
// created by elements within a workflow during simulation. To avoid
// unnecessary rewriting and duplication, the simulation essentially
// stores created files in a denormalised manner. Consequently,
// tooling is required in order to easily see which files were
// produced by which elements. This command is that tooling.
//
// For a workflow with an element of name MyElement, id=1, which has a
// Data named MyDataFile of type *File, where that *File value is non
// empty and has a Name property of TheFileName.txt, this tool will
// write that file content to
// outdir/1_MyElement/MyDataFile/TheFileName.txt.
//
// All *File and []*File Data values from all elements within the
// workflow are extracted. Log messages indicate progress.  Note that
// this only works if the indir and outdir of the simulation itself
// still exist, are accessible, and are unmodified.
package main
