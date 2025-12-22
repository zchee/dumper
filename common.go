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
	"fmt"
	"io"
	"math"
	"math/bits"
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

const (
	// ptrSize is the size of a pointer on the current arch.
	ptrSize = unsafe.Sizeof((*byte)(nil))
)

var (
	// offsetPtr, offsetScalar, and offsetFlag are the offsets for the
	// internal reflect.Value fields.  These values are valid before golang
	// commit ecccf07e7f9d which changed the format.  The are also valid
	// after commit 82f48826c6c7 which changed the format again to mirror
	// the original format.  Code in the init function updates these offsets
	// as necessary.
	offsetPtr    = uintptr(ptrSize)
	offsetScalar = uintptr(0)
	offsetFlag   = uintptr(ptrSize * 2)

	// flagKindWidth and flagKindShift indicate various bits that the
	// reflect package uses internally to track kind information.
	//
	// flagRO indicates whether or not the value field of a reflect.Value is
	// read-only.
	//
	// flagIndir indicates whether the value field of a reflect.Value is
	// the actual data or a pointer to the data.
	//
	// These values are valid before golang commit 90a7c3c86944 which
	// changed their positions.  Code in the init function updates these
	// flags as necessary.
	flagKindWidth = uintptr(5)
	flagKindShift = uintptr(flagKindWidth - 1)
	flagRO        = uintptr(1 << 0)
	flagIndir     = uintptr(1 << 1)
)

func init() {
	// Older versions of reflect.Value stored small integers directly in the
	// ptr field (which is named val in the older versions).  Versions
	// between commits ecccf07e7f9d and 82f48826c6c7 added a new field named
	// scalar for this purpose which unfortunately came before the flag
	// field, so the offset of the flag field is different for those
	// versions.
	//
	// This code constructs a new reflect.Value from a known small integer
	// and checks if the size of the reflect.Value struct indicates it has
	// the scalar field. When it does, the offsets are updated accordingly.
	vv := reflect.ValueOf(0xf00)
	if unsafe.Sizeof(vv) == (ptrSize * 4) {
		offsetScalar = ptrSize * 2
		offsetFlag = ptrSize * 3
	}

	// Commit 90a7c3c86944 changed the flag positions such that the low
	// order bits are the kind.  This code extracts the kind from the flags
	// field and ensures it's the correct type.  When it's not, the flag
	// order has been changed to the newer format, so the flags are updated
	// accordingly.
	upf := unsafe.Add(unsafe.Pointer(&vv), offsetFlag)
	upfv := *(*uintptr)(upf)
	flagKindMask := uintptr((1<<flagKindWidth - 1) << flagKindShift)
	if (upfv&flagKindMask)>>flagKindShift != uintptr(reflect.Int) {
		flagKindShift = 0
		flagRO = 1 << 5
		flagIndir = 1 << 6

		// Commit adf9b30e5594 modified the flags to separate the
		// flagRO flag into two bits which specifies whether or not the
		// field is embedded.  This causes flagIndir to move over a bit
		// and means that flagRO is the combination of either of the
		// original flagRO bit and the new bit.
		//
		// This code detects the change by extracting what used to be
		// the indirect bit to ensure it's set.  When it's not, the flag
		// order has been changed to the newer format, so the flags are
		// updated accordingly.
		if upfv&flagIndir == 0 {
			flagRO = 3 << 5
			flagIndir = 1 << 7
		}
	}

	for i := range hexByteTable {
		hexByteTable[i][0] = '0'
		hexByteTable[i][1] = 'x'
		hexByteTable[i][2] = hexDigits[i>>4]
		hexByteTable[i][3] = hexDigits[i&0x0f]
		hexByteTable[i][4] = ','
	}
}

// unsafeReflectValue converts the passed reflect.Value into a one that bypasses
// the typical safety restrictions preventing access to unaddressable and
// unexported data.  It works by digging the raw pointer to the underlying
// value out of the protected value and generating a new unprotected (unsafe)
// reflect.Value to it.
//
// This allows us to check for implementations of the Stringer and error
// interfaces to be used for pretty printing ordinarily unaddressable and
// inaccessible values such as unexported struct fields.
func unsafeReflectValue(v reflect.Value) (rv reflect.Value) {
	indirects := 1
	vt := v.Type()
	upv := unsafe.Add(unsafe.Pointer(&v), offsetPtr)
	rvf := *(*uintptr)(unsafe.Add(unsafe.Pointer(&v), offsetFlag))
	if rvf&flagIndir != 0 {
		vt = reflect.PointerTo(v.Type())
		indirects++
	} else if offsetScalar != 0 {
		// The value is in the scalar field when it's not one of the
		// reference types.
		switch vt.Kind() {
		case reflect.Uintptr:
		case reflect.Chan:
		case reflect.Func:
		case reflect.Map:
		case reflect.Pointer:
		case reflect.UnsafePointer:
		default:
			upv = unsafe.Add(unsafe.Pointer(&v), offsetScalar)
		}
	}

	pv := reflect.NewAt(vt, upv)
	rv = pv
	for i := 0; i < indirects; i++ {
		rv = rv.Elem()
	}
	return rv
}

// Some constants in the form of bytes to avoid string overhead.  This mirrors
// the technique used in the fmt package.
var (
	backQuoteBytes        = []byte("`")
	plusBytes             = []byte("+")
	trueBytes             = []byte("true")
	falseBytes            = []byte("false")
	interfaceBytes        = []byte("interface{}")
	commaSpaceBytes       = []byte(", ")
	commaNewlineBytes     = []byte(",\n")
	newlineBytes          = []byte("\n")
	openBraceBytes        = []byte("{")
	openBraceNewlineBytes = []byte("{\n")
	closeBraceBytes       = []byte("}")
	ampersandBytes        = []byte("&")
	colonSpaceBytes       = []byte(": ")
	spaceBytes            = []byte(" ")
	openParenBytes        = []byte("(")
	closeParenBytes       = []byte(")")
	nilBytes              = []byte("nil")
	hexZeroBytes          = []byte("0x")
	zeroBytes             = []byte("0")
	openCommentBytes      = []byte(" /*")
	closeCommentBytes     = []byte("*/ ")
	pointerChainBytes     = []byte("->")
	circularBytes         = []byte("(<already shown>)")
	invalidAngleBytes     = []byte("<invalid>")
	commentPrefixBytes    = []byte(" // |")
	commentSuffixBytes    = []byte("|\n")
)

// hexDigits is used to map a decimal value to a hex digit.
var hexDigits = "0123456789abcdef"

var hexByteTable [256][5]byte

// printBool outputs a boolean value as true or false to Writer w.
func printBool(w io.Writer, val bool) {
	if val {
		w.Write(trueBytes)
	} else {
		w.Write(falseBytes)
	}
}

// printInt outputs a signed integer value to Writer w.
func printInt(w io.Writer, scratch []byte, val int64, base int) {
	formatted := strconv.AppendInt(scratch[:0], val, base)
	w.Write(formatted)
}

// printUint outputs an unsigned integer value to Writer w.
func printUint(w io.Writer, scratch []byte, val uint64, base int) {
	formatted := strconv.AppendUint(scratch[:0], val, base)
	w.Write(formatted)
}

// printFloat outputs a floating point value using the specified precision,
// which is expected to be 32 or 64bit, to Writer w.
func printFloat(w io.Writer, scratch []byte, val float64, precision int, typeElided bool) {
	formatted := strconv.AppendFloat(scratch[:0], val, 'g', -1, precision)
	if typeElided && !math.IsInf(val, 0) && val == math.Floor(val) {
		formatted = append(formatted, '.', '0')
	}
	w.Write(formatted)
}

// printComplex outputs a complex value using the specified float precision
// for the real and imaginary parts to Writer w.
func printComplex(w io.Writer, scratch []byte, c complex128, floatPrecision int) {
	formatted := strconv.AppendFloat(scratch[:0], real(c), 'g', -1, floatPrecision)
	if imag(c) >= 0 {
		formatted = append(formatted, '+')
	}
	formatted = strconv.AppendFloat(formatted, imag(c), 'g', -1, floatPrecision)
	formatted = append(formatted, 'i')
	w.Write(formatted)
}

// hexDump is a modified 'hexdump -C'-like that returns a commented Go syntax
// byte slice or array.
func hexDump(w io.Writer, data []byte, indent []byte, width int, comment, addr bool) {
	if width <= 0 {
		width = 16 // This is the width used by hexdump -C, so it makes a reasonable default.
	}
	if len(data) == 0 {
		return
	}

	var commentBytes []byte
	if comment {
		commentBytes = make([]byte, width)
	}
	needsPadding := comment && len(data) > width && len(data)%width != 0

	addrWidth := 0
	if addr {
		addrWidth = (bits.Len(uint(len(data))) + 3) / 4
		if addrWidth == 0 {
			addrWidth = 1
		}
	}
	var addrBuf [32]byte
	var addrDigitsBuf [32]byte
	padCap := 0
	if needsPadding {
		remainder := len(data) % width
		slots := width - remainder
		switch {
		case slots <= 0:
			// Do nothing.
		case slots == 1:
			padCap = 6
		default:
			padCap = 12 + (slots-2)*6
		}
	}

	lineCap := len(indent)
	if addr {
		lineCap += 2 + addrWidth + 2
	}
	lineCap += width * 6
	if comment {
		lineCap += len(commentPrefixBytes) + len(commentSuffixBytes) + width
		if needsPadding {
			lineCap += padCap
		}
	} else {
		lineCap++
	}

	line := make([]byte, 0, lineCap)
	for offset := 0; offset < len(data); offset += width {
		line = line[:0]
		line = append(line, indent...)
		if addr {
			buf := addrBuf[:0]
			buf = append(buf, '0', 'x')
			digits := strconv.AppendUint(addrDigitsBuf[:0], uint64(offset), 16)
			if pad := addrWidth - len(digits); pad > 0 {
				for range pad {
					buf = append(buf, '0')
				}
			}
			buf = append(buf, digits...)
			buf = append(buf, ':', ' ')
			line = append(line, buf...)
		}

		lineLen := width
		if remaining := len(data) - offset; remaining < width {
			lineLen = remaining
		}

		for i := 0; i < lineLen; i++ {
			if i > 0 {
				line = append(line, ' ')
			}
			b := data[offset+i]
			line = append(line, hexByteTable[b][:]...)
			if comment {
				cb := b
				if cb < 32 || cb > 126 {
					cb = '.'
				}
				commentBytes[i] = cb
			}
		}

		if !comment {
			line = append(line, '\n')
			w.Write(line)
			continue
		}

		if lineLen == width {
			line = append(line, commentPrefixBytes...)
			line = append(line, commentBytes[:width]...)
			line = append(line, commentSuffixBytes...)
			w.Write(line)
			continue
		}

		if needsPadding {
			slots := width - lineLen
			switch slots {
			case 0:
				// Do nothing.
			case 1:
				line = append(line, ' ', '/', '*', ' ', '*', '/')
			default:
				line = append(line, ' ', '/', '*', ' ', ' ', ' ')
				for i := 0; i < slots-2; i++ {
					line = append(line, ' ', ' ', ' ', ' ', ' ', ' ')
				}
				line = append(line, ' ', ' ', ' ', ' ', '*', '/')
			}
		}

		line = append(line, commentPrefixBytes...)
		line = append(line, commentBytes[:lineLen]...)
		line = append(line, commentSuffixBytes...)
		w.Write(line)
	}
}

// printHexPtr outputs a uintptr formatted as hexadecimal with a leading '0x'
// prefix to Writer w.
func printHexPtr(w io.Writer, p uintptr, isPointer bool) {
	// Null pointer.
	num := uint64(p)
	if num == 0 {
		if isPointer {
			w.Write(nilBytes)
		} else {
			w.Write(zeroBytes)
		}
		return
	}

	// Max uint64 is 16 bytes in hex + 2 bytes for '0x' prefix.
	var buf [18]byte

	// It's simpler to construct the hex string right to left.
	base := uint64(16)
	i := len(buf) - 1
	for num >= base {
		buf[i] = hexDigits[num%base]
		num /= base
		i--
	}
	buf[i] = hexDigits[num]

	// Add '0x' prefix.
	i--
	buf[i] = 'x'
	i--
	buf[i] = '0'

	// Strip unused leading bytes.
	w.Write(buf[i:])
}

// mapSorter implements sort.Interface to allow a slice of reflect.Value
// elements to be sorted.
type mapSorter struct {
	keys []reflect.Value
	vals []reflect.Value
}

// Len returns the number of values in the slice.  It is part of the
// sort.Interface implementation.
func (s *mapSorter) Len() int {
	return len(s.keys)
}

// Swap swaps the values at the passed indices.  It is part of the
// sort.Interface implementation.
func (s *mapSorter) Swap(i, j int) {
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
	s.vals[i], s.vals[j] = s.vals[j], s.vals[i]
}

// less returns whether the a key/value should sort before the b key/value.
// It is used by valueSorter.Less as part of the sort.Interface implementation.
func less(kA, kB, vA, vB reflect.Value) bool {
	switch kA.Kind() {
	case reflect.Bool:
		return !kA.Bool() && kB.Bool()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return kA.Int() < kB.Int()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return kA.Uint() < kB.Uint()
	case reflect.Float32, reflect.Float64:
		if vA.IsValid() && vB.IsValid() && math.IsNaN(kA.Float()) && math.IsNaN(kB.Float()) {
			return less(vA, vB, reflect.Value{}, reflect.Value{})
		}
		return math.IsNaN(kA.Float()) || kA.Float() < kB.Float()
	case reflect.String:
		return kA.String() < kB.String()
	case reflect.Uintptr:
		return kA.Uint() < kB.Uint()
	case reflect.Array:
		// Compare the contents of both arrays.
		l := kA.Len()
		for i := range l {
			av := kA.Index(i)
			bv := kB.Index(i)
			if av.Interface() == bv.Interface() {
				continue
			}
			return less(av, bv, vA, vB)
		}
		return less(vA, vB, reflect.Value{}, reflect.Value{})
	}
	return fmt.Sprint(kA) < fmt.Sprint(kB)
}

// Less returns whether the value at index i should sort before the
// value at index j.  It is part of the sort.Interface implementation.
func (s *mapSorter) Less(i, j int) bool {
	return less(s.keys[i], s.keys[j], s.vals[i], s.vals[j])
}

// sortMapByKeyVals is a generic sort function for native types: int, uint, bool,
// float, string and uintptr.  Other inputs are sorted according to their
// Value.String() value to ensure display stability. Floating point values
// sort NaN before non-NaN values and NaN keys are ordered by their corresponding
// values.
func sortMapByKeyVals(keys, vals []reflect.Value) {
	if len(keys) != len(vals) {
		panic("invalid map key val slice pair")
	}
	if len(keys) == 0 {
		return
	}
	sort.Sort(&mapSorter{keys: keys, vals: vals})
}
