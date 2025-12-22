/*
 * Copyright (c) 2013 Dave Collins <dave@davec.name>
 * Copyright (c) 2015 Dan Kortschak <dan.kortschak@adelaide.edu.au>
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
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/zchee/dumper"
)

type benchChild struct {
	ID     int
	Name   string
	Values []float64
}

type benchData struct {
	ID       int
	Name     string
	Numbers  []int
	Bytes    []byte
	Children []benchChild
	Nested   map[string]any
}

func newBenchData() benchData {
	numbers := make([]int, 1024)
	for i := range numbers {
		numbers[i] = i * 3
	}

	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}

	children := make([]benchChild, 32)
	for i := range children {
		values := make([]float64, 32)
		for j := range values {
			values[j] = float64(i*j) * 0.5
		}
		children[i] = benchChild{
			ID:     i,
			Name:   "child",
			Values: values,
		}
	}

	return benchData{
		ID:       42,
		Name:     "bench",
		Numbers:  numbers,
		Bytes:    payload,
		Children: children,
		Nested: map[string]any{
			"numbers": numbers[:64],
			"bytes":   payload[:64],
			"meta": map[string]any{
				"active": true,
				"count":  128,
				"name":   "nested",
			},
			"mixed": []any{"a", 1, 2.5, true},
		},
	}
}

func TestNewBenchData(t *testing.T) {
	tests := map[string]struct {
		want struct {
			NumbersLen  int
			BytesLen    int
			ChildrenLen int
			NestedLen   int
		}
	}{
		"success: basic shape": {
			want: struct {
				NumbersLen  int
				BytesLen    int
				ChildrenLen int
				NestedLen   int
			}{
				NumbersLen:  1024,
				BytesLen:    256,
				ChildrenLen: 32,
				NestedLen:   4,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data := newBenchData()
			got := struct {
				NumbersLen  int
				BytesLen    int
				ChildrenLen int
				NestedLen   int
			}{
				NumbersLen:  len(data.Numbers),
				BytesLen:    len(data.Bytes),
				ChildrenLen: len(data.Children),
				NestedLen:   len(data.Nested),
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected bench data shape (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkSdump(b *testing.B) {
	data := newBenchData()
	defaultConfig := dumper.NewDefaultConfig()
	sortedConfig := dumper.NewDefaultConfig()
	sortedConfig.SortKeys = true
	noCommentConfig := dumper.NewDefaultConfig()
	noCommentConfig.CommentBytes = false
	addressConfig := dumper.NewDefaultConfig()
	addressConfig.AddressBytes = true

	benchmarks := map[string]struct {
		cfg   *dumper.ConfigState
		value any
	}{
		"default: nested struct": {
			cfg:   defaultConfig,
			value: data,
		},
		"default: numeric slice": {
			cfg:   defaultConfig,
			value: data.Numbers,
		},
		"default: byte slice": {
			cfg:   defaultConfig,
			value: data.Bytes,
		},
		"default: byte slice (no comment)": {
			cfg:   noCommentConfig,
			value: data.Bytes,
		},
		"default: byte slice (address)": {
			cfg:   addressConfig,
			value: data.Bytes,
		},
		"sorted: nested map": {
			cfg:   sortedConfig,
			value: data.Nested,
		},
	}

	for name, bench := range benchmarks {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			var sink string
			for b.Loop() {
				sink = bench.cfg.Sdump(bench.value)
			}
			if sink == "" {
				b.Fatalf("unexpected empty dump")
			}
		})
	}
}

func BenchmarkFdump(b *testing.B) {
	data := newBenchData()
	defaultConfig := dumper.NewDefaultConfig()
	noCommentConfig := dumper.NewDefaultConfig()
	noCommentConfig.CommentBytes = false
	addressConfig := dumper.NewDefaultConfig()
	addressConfig.AddressBytes = true

	benchmarks := map[string]struct {
		cfg   *dumper.ConfigState
		value any
	}{
		"default: nested struct": {
			cfg:   defaultConfig,
			value: data,
		},
		"default: byte slice": {
			cfg:   defaultConfig,
			value: data.Bytes,
		},
		"default: byte slice (no comment)": {
			cfg:   noCommentConfig,
			value: data.Bytes,
		},
		"default: byte slice (address)": {
			cfg:   addressConfig,
			value: data.Bytes,
		},
	}

	for name, bench := range benchmarks {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			var sink bytes.Buffer
			for b.Loop() {
				sink.Reset()
				bench.cfg.Fdump(&sink, bench.value)
			}
			if sink.Len() == 0 {
				b.Fatalf("unexpected empty dump")
			}
		})
	}
}
