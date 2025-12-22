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

//nolint:unused
var (
	ansiBlack         = []byte("\x1b[30m")
	ansiRed           = []byte("\x1b[31m")
	ansiGreen         = []byte("\x1b[32m")
	ansiYellow        = []byte("\x1b[33m")
	ansiBlue          = []byte("\x1b[34m")
	ansiMagenta       = []byte("\x1b[35m")
	ansiCyan          = []byte("\x1b[36m")
	ansiWhite         = []byte("\x1b[37m")
	ansiBoldBlack     = []byte("\x1b[30;1m")
	ansiBoldRed       = []byte("\x1b[31;1m")
	ansiBoldGreen     = []byte("\x1b[32;1m")
	ansiBoldYellow    = []byte("\x1b[33;1m")
	ansiBoldBlue      = []byte("\x1b[34;1m")
	ansiBoldMagenta   = []byte("\x1b[35;1m")
	ansiBoldCyan      = []byte("\x1b[36;1m")
	ansiBoldWhite     = []byte("\x1b[37;1m")
	ansiBrightBlack   = []byte("\x1b[90m")
	ansiBrightRed     = []byte("\x1b[91m")
	ansiBrightGreen   = []byte("\x1b[92m")
	ansiBrightYellow  = []byte("\x1b[93m")
	ansiBrightBlue    = []byte("\x1b[94m")
	ansiBrightMagenta = []byte("\x1b[95m")
	ansiBrightCyan    = []byte("\x1b[96m")
	ansiBrightWhite   = []byte("\x1b[97m")
)

var kindColor = [kindColorCount][]byte{
	reflect.Invalid:       ansiBrightBlack,
	reflect.Bool:          ansiYellow,
	reflect.Int:           ansiMagenta,
	reflect.Int8:          ansiMagenta,
	reflect.Int16:         ansiMagenta,
	reflect.Int32:         ansiMagenta,
	reflect.Int64:         ansiMagenta,
	reflect.Uint:          ansiMagenta,
	reflect.Uint8:         ansiMagenta,
	reflect.Uint16:        ansiMagenta,
	reflect.Uint32:        ansiMagenta,
	reflect.Uint64:        ansiMagenta,
	reflect.Uintptr:       ansiMagenta,
	reflect.Float32:       ansiBoldMagenta,
	reflect.Float64:       ansiBoldMagenta,
	reflect.Complex64:     ansiRed,
	reflect.Complex128:    ansiRed,
	reflect.Array:         ansiBlue,
	reflect.Chan:          ansiBrightMagenta,
	reflect.Func:          ansiBrightYellow,
	reflect.Interface:     ansiBrightBlack,
	reflect.Map:           ansiBrightCyan,
	reflect.Pointer:       ansiBoldBlue,
	reflect.Slice:         ansiBoldBlue,
	reflect.String:        ansiGreen,
	reflect.Struct:        ansiWhite,
	reflect.UnsafePointer: ansiBrightRed,
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
