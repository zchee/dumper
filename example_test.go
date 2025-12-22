/*
 * Copyright (c) 2013 Dave Collins <dave@davec.name>
 * Copyright (c) 2015 Dan Kortschak <dan.kortschak@adelaide.edu.au>
 * Copyright (c) 2025 Koichi Shiraishi <zchee.io@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package dumper_test

import (
	"fmt"

	"github.com/zchee/dumper"
)

type Flag int

const (
	flagOne Flag = iota
	flagTwo
)

var flagStrings = map[Flag]string{
	flagOne: "flagOne",
	flagTwo: "flagTwo",
}

func (f Flag) String() string {
	if s, ok := flagStrings[f]; ok {
		return s
	}
	return fmt.Sprintf("Unknown flag (%d)", int(f))
}

type Bar struct {
	flag Flag
	data uintptr
}

type Foo struct {
	unexportedField Bar
	ExportedField   map[any]any
}

// This example demonstrates how to use Dump to dump variables to stdout.
func ExampleDump() {
	// The following package level declarations are assumed for this example:
	/*
		type Flag int

		const (
			flagOne Flag = iota
			flagTwo
		)

		var flagStrings = map[Flag]string{
			flagOne: "flagOne",
			flagTwo: "flagTwo",
		}

		func (f Flag) String() string {
			if s, ok := flagStrings[f]; ok {
				return s
			}
			return fmt.Sprintf("Unknown flag (%d)", int(f))
		}

		type Bar struct {
			flag Flag
			data uintptr
		}

		type Foo struct {
			unexportedField Bar
			ExportedField   map[interface{}]interface{}
		}
	*/

	// Setup some sample data structures for the example.
	bar := Bar{Flag(flagTwo), uintptr(0)}
	s1 := Foo{bar, map[any]any{"one": true}}
	f := Flag(5)
	b := []byte{
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30,
		0x31, 0x32,
	}

	prev := dumper.Config.EnableColor
	dumper.Config.EnableColor = true
	defer func() {
		dumper.Config.EnableColor = prev
	}()

	// Dump!
	dumper.Dump([]any{s1, f, b})

	// Output:
	//
	// [34;1m[]interface{}[0m[34;1m{
	// [0m [37mdumper_test.Foo[0m[37m{
	// [0m  unexportedField: [37mdumper_test.Bar[0m[37m{
	// [0m   flag: [35mdumper_test.Flag[0m([35m1[0m),
	//    data: [35muintptr[0m([35m0[0m),
	//   [37m}[0m,
	//   ExportedField: [96mmap[interface{}]interface{}[0m[96m{
	// [0m   [32mstring[0m([32m"one"[0m): [33mbool[0m([33mtrue[0m),
	//   [96m}[0m,
	//  [37m}[0m,
	//  [35mdumper_test.Flag[0m([35m5[0m),
	//  [34;1m[]uint8[0m[34;1m{
	// [0m[34;1m  0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, // |............... |
	//   0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, // |!"#$%&'()*+,-./0|
	//   0x31, 0x32, /*                                                                               */ // |12|
	// [0m [34;1m}[0m,
	// [34;1m}[0m
}

// This example demonstrates how to use a ConfigState.
func ExampleConfigState() {
	// Modify the indent level of the ConfigState only.  The global
	// configuration is not modified.
	scs := dumper.ConfigState{Indent: "\t", EnableColor: true}

	// Output using the ConfigState instance.
	v := map[string]int{"one": 1}
	scs.Dump(v)

	// Output:
	//
	// [96mmap[string]int[0m[96m{
	// [0m	[32mstring[0m([32m"one"[0m): [35mint[0m([35m1[0m),
	// [96m}[0m
}

// This example demonstrates how to use a Quoting strategy.
func ExampleConfigState_Quoting() {
	scs := dumper.ConfigState{
		Indent:      "\t",
		ElideType:   true,
		SortKeys:    true,
		EnableColor: true,

		// Avoid escape sequences when present and force
		// use of backquotes even when the complete string is
		// not backquotable.
		Quoting: dumper.AvoidEscapes | dumper.Force,
	}

	v := map[string]string{
		"1. one":              "this\ntext\nspans\nlines\n",
		"2. two":              "this text doesn't",
		"3.\nt\nh\nr\ne\ne\n": "vertical key",
		"4. four":             "contains \\backslashes\\ and `backquotes`",
	}
	scs.Dump(v)

	// Output:
	//
	// [96mmap[string]string[0m[96m{
	// [0m	[32m"1. one"[0m: [32m`this
	// text
	// spans
	// lines
	// `[0m,
	// 	[32m"2. two"[0m: [32m"this text doesn't"[0m,
	// 	[32m`3.
	// t
	// h
	// r
	// e
	// e
	// `[0m: [32m"vertical key"[0m,
	// 	[32m"4. four"[0m: [32m`contains \backslashes\ and `+"`"+`backquotes`+"`"[0m,
	// [96m}[0m
}

// This example demonstrates how to use ConfigState.Dump to dump variables to
// stdout
func ExampleConfigState_Dump() {
	// See the top-level Dump example for details on the types used in this
	// example.

	// Create two ConfigState instances with different indentation.
	scs := dumper.ConfigState{Indent: "\t", EnableColor: true}
	scs2 := dumper.ConfigState{Indent: " ", EnableColor: true}

	// Setup some sample data structures for the example.
	bar := Bar{Flag(flagTwo), uintptr(0)}
	s1 := Foo{bar, map[any]any{"one": true}}

	// Dump using the ConfigState instances.
	scs.Dump(s1)
	scs2.Dump(s1)

	// Output:
	//
	// [37mdumper_test.Foo[0m[37m{
	// [0m	unexportedField: [37mdumper_test.Bar[0m[37m{
	// [0m		flag: [35mdumper_test.Flag[0m([35m1[0m),
	// 		data: [35muintptr[0m([35m0[0m),
	// 	[37m}[0m,
	// 	ExportedField: [96mmap[interface{}]interface{}[0m[96m{
	// [0m		[32mstring[0m([32m"one"[0m): [33mbool[0m([33mtrue[0m),
	// 	[96m}[0m,
	// [37m}[0m
	// [37mdumper_test.Foo[0m[37m{
	// [0m unexportedField: [37mdumper_test.Bar[0m[37m{
	// [0m  flag: [35mdumper_test.Flag[0m([35m1[0m),
	//   data: [35muintptr[0m([35m0[0m),
	//  [37m}[0m,
	//  ExportedField: [96mmap[interface{}]interface{}[0m[96m{
	// [0m  [32mstring[0m([32m"one"[0m): [33mbool[0m([33mtrue[0m),
	//  [96m}[0m,
	// [37m}[0m
}
