package data

import "reflect"

// Code generated by gen.py. DO NOT EDIT.

// registry for well-known scalar types

var typeFloat64 reflect.Type = reflect.TypeOf(float64(0))
var typeInt64 reflect.Type = reflect.TypeOf(int64(0))
var typeString reflect.Type = reflect.TypeOf("")
var typeBool reflect.Type = reflect.TypeOf(false)
var typeTimestampMillis reflect.Type = reflect.TypeOf(TimestampMillis(0))
var typeTimestampMicros reflect.Type = reflect.TypeOf(TimestampMicros(0))

var typeSupport map[reflect.Type]string = map[reflect.Type]string{
	typeFloat64:         "float64",
	typeInt64:           "int64",
	typeString:          "string",
	typeBool:            "bool",
	typeTimestampMillis: "TimestampMillis",
	typeTimestampMicros: "TimestampMicros",
}

var typeSupportByName map[string]reflect.Type = map[string]reflect.Type{
	"float64":         typeFloat64,
	"int64":           typeInt64,
	"string":          typeString,
	"bool":            typeBool,
	"TimestampMillis": typeTimestampMillis,
	"TimestampMicros": typeTimestampMicros,
}
