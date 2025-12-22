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
	"bytes"
	"strings"
	"testing"

	"github.com/zchee/dumper"
)

func TestFdumpColorized(t *testing.T) {
	buf := new(bytes.Buffer)
	dumper.Fdump(buf, 42)
	out := buf.String()
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI sequences in Fdump output: %q", out)
	}
	if stripANSI(out) == out {
		t.Fatalf("expected ANSI sequences to change output")
	}
}

func TestFdumpDisableColor(t *testing.T) {
	cfg := dumper.ConfigState{DisableColor: true}
	buf := new(bytes.Buffer)
	cfg.Fdump(buf, 42)
	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("unexpected ANSI sequences in disabled output: %q", out)
	}
}

func TestSdumpNoColor(t *testing.T) {
	out := dumper.Sdump(42)
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("unexpected ANSI sequences in Sdump output: %q", out)
	}
}
