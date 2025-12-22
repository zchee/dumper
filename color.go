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

package dumper

import (
	"io"
	"reflect"
)

const kindColorCount = int(reflect.UnsafePointer) + 1

var ansiReset = []byte("\x1b[0m")

var kindColor = [kindColorCount][]byte{
	reflect.Invalid:       []byte("\x1b[38;5;240m"),
	reflect.Bool:          []byte("\x1b[38;5;46m"),
	reflect.Int:           []byte("\x1b[38;5;208m"),
	reflect.Int8:          []byte("\x1b[38;5;202m"),
	reflect.Int16:         []byte("\x1b[38;5;214m"),
	reflect.Int32:         []byte("\x1b[38;5;220m"),
	reflect.Int64:         []byte("\x1b[38;5;226m"),
	reflect.Uint:          []byte("\x1b[38;5;118m"),
	reflect.Uint8:         []byte("\x1b[38;5;82m"),
	reflect.Uint16:        []byte("\x1b[38;5;75m"),
	reflect.Uint32:        []byte("\x1b[38;5;51m"),
	reflect.Uint64:        []byte("\x1b[38;5;39m"),
	reflect.Uintptr:       []byte("\x1b[38;5;33m"),
	reflect.Float32:       []byte("\x1b[38;5;129m"),
	reflect.Float64:       []byte("\x1b[38;5;93m"),
	reflect.Complex64:     []byte("\x1b[38;5;161m"),
	reflect.Complex128:    []byte("\x1b[38;5;198m"),
	reflect.Array:         []byte("\x1b[38;5;69m"),
	reflect.Chan:          []byte("\x1b[38;5;105m"),
	reflect.Func:          []byte("\x1b[38;5;135m"),
	reflect.Interface:     []byte("\x1b[38;5;244m"),
	reflect.Map:           []byte("\x1b[38;5;31m"),
	reflect.Pointer:       []byte("\x1b[38;5;196m"),
	reflect.Slice:         []byte("\x1b[38;5;45m"),
	reflect.String:        []byte("\x1b[38;5;34m"),
	reflect.Struct:        []byte("\x1b[38;5;141m"),
	reflect.UnsafePointer: []byte("\x1b[38;5;160m"),
}

func colorForKind(kind reflect.Kind) []byte {
	if int(kind) >= len(kindColor) {
		return nil
	}
	return kindColor[kind]
}

func writeColor(w io.Writer, kind reflect.Kind, colorize bool, write func()) {
	if !colorize {
		write()
		return
	}
	color := colorForKind(kind)
	if len(color) == 0 {
		write()
		return
	}
	w.Write(color)
	write()
	w.Write(ansiReset)
}
