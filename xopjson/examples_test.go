// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson_test

import "github.com/muir/xop-go/xopconst"

type AnyObject struct {
	I int
	S string
	A []string
	P *AnyObject
}

var (
	ExampleMetadataSingleAny   = xopconst.Make{Key: "s-any", Namespace: "test"}.AnyAttribute(AnyObject{})
	ExampleMetadataLockedAny   = xopconst.Make{Key: "l-any", Locked: true, Namespace: "test"}.AnyAttribute(AnyObject{})
	ExampleMetadataMultipleAny = xopconst.Make{Key: "m-any", Multiple: true, Namespace: "test"}.AnyAttribute(AnyObject{})
	ExampleMetadataDistinctAny = xopconst.Make{Key: "d-any", Multiple: true, Distinct: true, Namespace: "test"}.AnyAttribute(AnyObject{})
)

var (
	ExampleMetadataSingleEnum   = xopconst.Make{Key: "s-enum", Namespace: "test"}.EmbeddedEnumAttribute()
	ExampleMetadataLockedEnum   = xopconst.Make{Key: "l-enum", Locked: true, Namespace: "test"}.EmbeddedEnumAttribute()
	ExampleMetadataMultipleEnum = xopconst.Make{Key: "m-enum", Multiple: true, Namespace: "test"}.EmbeddedEnumAttribute()
	ExampleMetadataDistinctEnum = xopconst.Make{Key: "d-enum", Multiple: true, Distinct: true, Namespace: "test"}.EmbeddedEnumAttribute()
)

var (
	SingleEnumOne   = ExampleMetadataSingleEnum.Iota("one")
	SingleEnumTwo   = ExampleMetadataSingleEnum.Iota("two")
	SingleEnumThree = ExampleMetadataSingleEnum.Iota("Three")

	LockedEnumOne   = ExampleMetadataLockedEnum.Iota("one")
	LockedEnumTwo   = ExampleMetadataLockedEnum.Iota("two")
	LockedEnumThree = ExampleMetadataLockedEnum.Iota("Three")

	MultipleEnumOne   = ExampleMetadataMultipleEnum.Iota("one")
	MultipleEnumTwo   = ExampleMetadataMultipleEnum.Iota("two")
	MultipleEnumThree = ExampleMetadataMultipleEnum.Iota("Three")

	DistinctEnumOne   = ExampleMetadataDistinctEnum.Iota("one")
	DistinctEnumTwo   = ExampleMetadataDistinctEnum.Iota("two")
	DistinctEnumThree = ExampleMetadataDistinctEnum.Iota("Three")
)

// TODO: why the skips?
var ExampleMetadataSingleBool = xopconst.Make{Key: "s-bool", Namespace: "test"}.BoolAttribute()

var (
	ExampleMetadataLockedBool       = xopconst.Make{Key: "l-bool", Locked: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataMultipleBool     = xopconst.Make{Key: "m-bool", Multiple: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataDistinctBool     = xopconst.Make{Key: "d-bool", Multiple: true, Distinct: true, Namespace: "test"}.BoolAttribute()
	ExampleMetadataSingleDuration   = xopconst.Make{Key: "s-time.Duration", Namespace: "test"}.DurationAttribute()
	ExampleMetadataLockedDuration   = xopconst.Make{Key: "l-time.Duration", Locked: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataMultipleDuration = xopconst.Make{Key: "m-time.Duration", Multiple: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataDistinctDuration = xopconst.Make{Key: "d-time.Duration", Multiple: true, Distinct: true, Namespace: "test"}.DurationAttribute()
	ExampleMetadataSingleInt        = xopconst.Make{Key: "s-int", Namespace: "test"}.IntAttribute()
	ExampleMetadataLockedInt        = xopconst.Make{Key: "l-int", Locked: true, Namespace: "test"}.IntAttribute()
	ExampleMetadataMultipleInt      = xopconst.Make{Key: "m-int", Multiple: true, Namespace: "test"}.IntAttribute()
	ExampleMetadataDistinctInt      = xopconst.Make{Key: "d-int", Multiple: true, Distinct: true, Namespace: "test"}.IntAttribute()
	ExampleMetadataSingleInt16      = xopconst.Make{Key: "s-int16", Namespace: "test"}.Int16Attribute()
	ExampleMetadataLockedInt16      = xopconst.Make{Key: "l-int16", Locked: true, Namespace: "test"}.Int16Attribute()
	ExampleMetadataMultipleInt16    = xopconst.Make{Key: "m-int16", Multiple: true, Namespace: "test"}.Int16Attribute()
	ExampleMetadataDistinctInt16    = xopconst.Make{Key: "d-int16", Multiple: true, Distinct: true, Namespace: "test"}.Int16Attribute()
	ExampleMetadataSingleInt32      = xopconst.Make{Key: "s-int32", Namespace: "test"}.Int32Attribute()
	ExampleMetadataLockedInt32      = xopconst.Make{Key: "l-int32", Locked: true, Namespace: "test"}.Int32Attribute()
	ExampleMetadataMultipleInt32    = xopconst.Make{Key: "m-int32", Multiple: true, Namespace: "test"}.Int32Attribute()
	ExampleMetadataDistinctInt32    = xopconst.Make{Key: "d-int32", Multiple: true, Distinct: true, Namespace: "test"}.Int32Attribute()
	ExampleMetadataSingleInt64      = xopconst.Make{Key: "s-int64", Namespace: "test"}.Int64Attribute()
	ExampleMetadataLockedInt64      = xopconst.Make{Key: "l-int64", Locked: true, Namespace: "test"}.Int64Attribute()
	ExampleMetadataMultipleInt64    = xopconst.Make{Key: "m-int64", Multiple: true, Namespace: "test"}.Int64Attribute()
	ExampleMetadataDistinctInt64    = xopconst.Make{Key: "d-int64", Multiple: true, Distinct: true, Namespace: "test"}.Int64Attribute()
	ExampleMetadataSingleInt8       = xopconst.Make{Key: "s-int8", Namespace: "test"}.Int8Attribute()
	ExampleMetadataLockedInt8       = xopconst.Make{Key: "l-int8", Locked: true, Namespace: "test"}.Int8Attribute()
	ExampleMetadataMultipleInt8     = xopconst.Make{Key: "m-int8", Multiple: true, Namespace: "test"}.Int8Attribute()
	ExampleMetadataDistinctInt8     = xopconst.Make{Key: "d-int8", Multiple: true, Distinct: true, Namespace: "test"}.Int8Attribute()
	ExampleMetadataSingleLink       = xopconst.Make{Key: "s-trace.Trace", Namespace: "test"}.LinkAttribute()
	ExampleMetadataLockedLink       = xopconst.Make{Key: "l-trace.Trace", Locked: true, Namespace: "test"}.LinkAttribute()
	ExampleMetadataMultipleLink     = xopconst.Make{Key: "m-trace.Trace", Multiple: true, Namespace: "test"}.LinkAttribute()
	ExampleMetadataDistinctLink     = xopconst.Make{Key: "d-trace.Trace", Multiple: true, Distinct: true, Namespace: "test"}.LinkAttribute()
	ExampleMetadataSingleString     = xopconst.Make{Key: "s-string", Namespace: "test"}.StringAttribute()
	ExampleMetadataLockedString     = xopconst.Make{Key: "l-string", Locked: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataMultipleString   = xopconst.Make{Key: "m-string", Multiple: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataDistinctString   = xopconst.Make{Key: "d-string", Multiple: true, Distinct: true, Namespace: "test"}.StringAttribute()
	ExampleMetadataSingleTime       = xopconst.Make{Key: "s-time.Time", Namespace: "test"}.TimeAttribute()
	ExampleMetadataLockedTime       = xopconst.Make{Key: "l-time.Time", Locked: true, Namespace: "test"}.TimeAttribute()
	ExampleMetadataMultipleTime     = xopconst.Make{Key: "m-time.Time", Multiple: true, Namespace: "test"}.TimeAttribute()
	ExampleMetadataDistinctTime     = xopconst.Make{Key: "d-time.Time", Multiple: true, Distinct: true, Namespace: "test"}.TimeAttribute()
)
