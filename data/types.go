package data

// TODO: the idea is to accumulate in types.go all the information about supported data types
// - and maybe even use it while code generation (requires rewriting code generation from Python to Go)

// Time and timestamp data types - in order to distinguish times and timestamps read from Parquet or created in Go code from ordinary integers

// TimeMillis is time measured in ms
type TimeMillis int32

// TimeMicros is time measured in us
type TimeMicros int64

// TimestampMillis is timestamp measured in ms
type TimestampMillis int64

// TimestampMicros is timestamp measured in us
type TimestampMicros int64
