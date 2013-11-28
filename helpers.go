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

// escapeTable is a table of offsets into the escape substTable
// for each ASCII character.  Zero represents no substitution.
var escapeTable = [...]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 1, 0, 0, 0, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 5,
}

var substTable = [...][]byte{
	{'&', 'q', 'u', 'o', 't', ';'}, // 1 "
	{'&', 'a', 'm', 'p', ';'},      // 2 &
	{'&', 'a', 'p', 'o', 's', ';'}, // 3 '
	{'&', 'l', 't', ';'},           // 4 <
	{'&', 'g', 't', ';'},           // 5 >
}

// escape generates an escaped XML string.
func escape(s string) string {
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if c := s[i]; int(c) < len(escapeTable) && escapeTable[c] > 0 {
			buf = append(buf, substTable[escapeTable[c]-1]...)
		} else {
			buf = append(buf, c)
		}
	}
	switch {
	case len(s) == len(buf):
		return s
	default:
		return string(buf)
	}
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

// crSpaces returns a carriage return followed by n spaces. It's used
// to generate XML indentations.
func crSpaces(n int) string {
	if n+1 > len(crsp) {
		buf := make([]byte, n+1)
		buf[0] = '\n'
		for i := 1; i < n+1; i++ {
			buf[i] = ' '
		}
		return string(buf)
	} else {
		return crsp[:n+1]
	}
}
