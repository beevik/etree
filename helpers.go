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
func escape(b []byte) []byte {
    buf := make([]byte, 0, len(b))
    for _, c := range b {
        subst := escapeTable[c]
        if subst > 0 {
            buf = append(buf, substTable[subst-1]...)
        } else {
            buf = append(buf, c)
        }
    }
    return buf
}

// copyBytes makes a copy of a byte slice.
func copyBytes(b []byte) []byte {
    c := make([]byte, len(b))
    copy(c, b)
    return c
}

// isWhitespace returns true if the byte slice contains only
// whitespace characters.
func isWhitespace(b []byte) bool {
    for _, c := range b {
        if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
            return false
        }
    }
    return true
}

// isEqual compares a byte slice and a string and returns true
// if they are equal.
func isEqual(b []byte, s string) bool {
    if len(b) != len(s) {
        return false
    }
    for i := 0; i < len(s); i++ {
        if b[i] != s[i] {
            return false
        }
    }
    return true
}

var crsp = []byte{'\n',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
    ' ', ' ', ' ', ' ',
}

// spaces returns a carriage return followed by n spaces. It's used
// to generate XML indentations.
func crSpaces(n int) []byte {
    if n+1 > len(crsp) {
        buf := make([]byte, n+1)
        buf[0] = '\n'
        for i := 1; i < n+1; i++ {
            buf[i] = ' '
        }
        return buf
    } else {
        return crsp[:n+1]
    }
}
