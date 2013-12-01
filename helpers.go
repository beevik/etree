/*
 * Copyright 2013 Brett Vickers. All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 *    1. Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *
 *    2. Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY COPYRIGHT HOLDER ``AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY
 * OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package etree

import (
	"io"
	"strings"
)

// An element stack is a simple stack of elements.
type elementStack []*Element

func (s *elementStack) empty() bool {
	return len(*s) == 0
}

func (s *elementStack) push(e *Element) {
	*s = append(*s, e)
}

func (s *elementStack) pop() *Element {
	e := (*s)[len(*s)-1]
	(*s)[len(*s)-1] = nil
	*s = (*s)[:len(*s)-1]
	return e
}

func (s *elementStack) peek() *Element {
	return (*s)[len(*s)-1]
}

// countReader implements a proxy reader that counts the number of
// bytes read from its encapsulated reader.
type countReader struct {
	r     io.Reader
	bytes int64
}

func newCountReader(r io.Reader) *countReader {
	return &countReader{r: r}
}

func (cr *countReader) Read(p []byte) (n int, err error) {
	b, err := cr.r.Read(p)
	cr.bytes += int64(b)
	return b, err
}

// countWriter implements a proxy writer that counts the number of
// bytes written by its encapsulated writer.
type countWriter struct {
	w     io.Writer
	bytes int64
}

func newCountWriter(w io.Writer) *countWriter {
	return &countWriter{w: w}
}

func (cw *countWriter) Write(p []byte) (n int, err error) {
	b, err := cw.w.Write(p)
	cw.bytes += int64(b)
	return b, err
}

var xmlReplacer = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"&", "&amp;",
	"'", "&apos;",
	`"`, "&quot",
)

// escape generates an escaped XML string.
func escape(s string) string {
	return xmlReplacer.Replace(s)
}

// isWhitespace returns true if the byte slice contains only
// whitespace characters.
func isWhitespace(s string) bool {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

var crsp = "\n                                                                                "
var crtab = "\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"

// crIndent returns a carriage return followed by n indent characters.
// The indent characters come from the source string.
func crIndent(n int, source string) string {
	switch {
	case n < 0:
		return source[:1]
	case n+1 > len(source):
		buf := make([]byte, n+1)
		buf[0] = '\n'
		for i := 1; i < n+1; i++ {
			buf[i] = ' '
		}
		return string(buf)
	default:
		return source[:n+1]
	}
}

// nextIndex returns the index of the next occurrence of sep in s,
// starting from offset.  It returns -1 if the sep string is not found.
func nextIndex(s, sep string, offset int) int {
	switch i := strings.Index(s[offset:], sep); i {
	case -1:
		return -1
	default:
		return offset + i
	}
}

// isInteger returns true if the string s contains an integer.
func isInteger(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && !(i == 0 && s[i] == '-') {
			return false
		}
	}
	return true
}
