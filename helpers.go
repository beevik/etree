package etree

// escapeTable is a table of offsets into the escape substTable
// for each ASCII character.  Zero represents no substitution.
var escapeTable = [...]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 1, 0, 0, 0, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 5, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

var substTable = [...][]byte{
	{'&', 'q', 'u', 'o', 't', ';'}, // 1
	{'&', 'a', 'm', 'p', ';'},      // 2
	{'&', 'a', 'p', 'o', 's', ';'}, // 3
	{'&', 'l', 't', ';'},           // 4
	{'&', 'g', 't', ';'},           // 5
}

// escape generates an escaped XML string.
func escape(s string) string {
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		subst := escapeTable[c]
		if subst > 0 {
			buf = append(buf, substTable[subst-1]...)
		} else {
			buf = append(buf, c)
		}
	}
	return string(buf)
}

// isWhitespace returns true if the byte slice contains only
// whitespace characters.
func isWhitespace(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

var crsp = "\n                                                                                "

// spaces returns a carriage return followed by n spaces. It's used
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
