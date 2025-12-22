dumper
=====

[![Test](https://github.com/zchee/dumper/actions/workflows/ci.yaml/badge.svg)](https://github.com/zchee/dumper/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/zchee/dumper/graph/badge.svg?token=9AM2DfBEX6)](https://codecov.io/gh/zchee/dumper)
[![Go Reference](https://pkg.go.dev/badge/github.com/zchee/dumper.svg)](https://pkg.go.dev/github.com/zchee/dumper)

dumper is a fork of the outstanding [go-spew tool](https://github.com/davecgh/go-spew).
Where go-spew is an aid for debugging, providing annotation of dumped datastructures,
dumper is a tool for taking snapshots of data structures to include in tests or other
code. An dumper dump will not construct cyclic structure literals and a number of
pseudo-code representations of pointer-based structures will require subsequent
processing.

A comprehensive suite of tests with near 100% test coverage is provided to ensure
proper functionality. dumper is licensed under the liberal ISC license, so it may
be used in open source or commercial projects.

## Installation

```bash
go get github.com/zchee/dumper@latest
```

## Quick Start

To dump a variable with full newlines, indentation, type, and pointer
information use Dump, Fdump, or Sdump:

```go
dumper.Dump(myVar1)
dumper.Fdump(someWriter, myVar1)
str := dumper.Sdump(myVar1)
```

## Sample Dump Output

```go
main.Foo{
 unexportedField: &main.Bar{
  flag: main.Flag(1),
  data: uintptr(0),
 },
 ExportedField: map[interface{}]interface{}{
  string("one"): bool(true),
 },
}
```

```go
[]uint8{
 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, // |........|
 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, // |....... |
 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, // |!"#$%&'(|
 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, // |)*+,-./0|
 0x31, 0x32, /*                               */ // |12|
}
```

## Configuration Options

Configuration of dumper is handled by fields in the ConfigState type. For
convenience, all of the top-level functions use a global state available via the
dumper.Config global.

It is also possible to create a ConfigState instance that provides methods
equivalent to the top-level functions. This allows concurrent configuration
options. See the ConfigState documentation for more details.

```go
* Indent
	String to use for each indentation level for Dump functions.
	It is a single space by default.  A popular alternative is "\t".

* NumericWidth
	NumericWidth specifies the number of columns to use when dumping
	a numeric slice or array (including bool). Zero specifies all entries
	on one line.

* StringWidth
	StringWidth specifies the number of columns to use when dumping
	a string slice or array. Zero specifies all entries on one line.

* BytesWidth
	Number of byte columns to use when dumping byte slices and arrays.

* CommentBytes
	Specifies whether ASCII comment annotations are attached to byte
	slice and array dumps.

* CommentPointers
	CommentPointers specifies whether pointer information will be added
	as comments.

* IgnoreUnexported
	Specifies that unexported fields should be ignored.

* ElideType
	ElideType specifies that type information defined by context should
	not be printed in a dump.

* OmitZero specifies that zero values should not be printed in a dump.

* SortKeys
	Specifies map keys should be sorted before being printed. Use
	this to have a more deterministic, diffable output.  Note that
	only native types (bool, int, uint, floats, uintptr and string)
	are supported with other types sorted according to the
	reflect.Value.String() output which guarantees display stability.
	Natural map order is used by default.
```

## License

dumper is licensed under the liberal ISC License.
