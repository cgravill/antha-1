package wtype

// FileSeries is a datatype representing a set of files for data processing usecases:
// - the files are related to the same entity (likely to the same device);
// - the files implicitly form a single dataset - i.e. they likely have the same data schema;
// A parameter of type FileSeries signals to the UI that it should be set using a special query dialog (rather than manually filling the list of files).
type FileSeries struct {
	Files []File
}
