package etree

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

type EscapeMode byte

const (
	EscapeNormal EscapeMode = iota
	EscapeCanonicalText
	EscapeCanonicalAttr
)

type XmlWriter interface {
	WriteTo(*Document, io.Writer) (int64, error)
	Indent(*Document, int)
	IndentTabs(*Document)
	Clone() XmlWriter
}

type StringEscaper interface {
	WriteEscape(*bufio.Writer, string, EscapeMode)
}

type StandardWriter struct {
	StringEscaper
	Settings WriteSettings
}

func NewStandardXmlWriter() *StandardWriter {
	return &StandardWriter{StringEscaper: &StandardEscaper{}}
}

func (sw *StandardWriter) WriteTo(d *Document, w io.Writer) (int64, error) {
	cw := newCountWriter(w)
	b := bufio.NewWriter(cw)
	for _, c := range d.Child {
		sw.WriteToken(c, b)
	}
	err, n := b.Flush(), cw.bytes
	return n, err
}

func (sw *StandardWriter) Clone() XmlWriter {
	return &StandardWriter{Settings: sw.Settings.dup(), StringEscaper: sw.StringEscaper}
}

func (sw *StandardWriter) Indent(doc *Document, spaces int) {
	var indent indentFunc
	switch {
	case spaces < 0:
		indent = func(depth int) string { return "" }
	case sw.Settings.UseCRLF:
		indent = func(depth int) string { return indentCRLF(depth*spaces, indentSpaces) }
	default:
		indent = func(depth int) string { return indentLF(depth*spaces, indentSpaces) }
	}
	doc.Element.indent(0, indent)
}

func (sw *StandardWriter) IndentTabs(doc *Document) {
	var indent indentFunc
	switch sw.Settings.UseCRLF {
	case true:
		indent = func(depth int) string { return indentCRLF(depth, indentTabs) }
	default:
		indent = func(depth int) string { return indentLF(depth, indentTabs) }
	}
	doc.Element.indent(0, indent)
}

func (sw *StandardWriter) WriteToken(token Token, w *bufio.Writer) {
	switch t := token.(type) {
	case *Comment:
		sw.WriteComment(t, w)
	case *Directive:
		sw.WriteDirective(t, w)
	case *Element:
		sw.WriteElement(t, w)
	case *CharData:
		sw.WriteCharData(t, w)
	case *ProcInst:
		sw.WriteProcInst(t, w)
	default:
		panic(fmt.Errorf("unsupported token %v", t))
	}
}

func (*StandardWriter) WriteComment(c *Comment, w *bufio.Writer) {
	w.WriteString("<!--")
	w.WriteString(c.Data)
	w.WriteString("-->")
}

func (*StandardWriter) WriteDirective(d *Directive, w *bufio.Writer) {
	w.WriteString("<!")
	w.WriteString(d.Data)
	w.WriteString(">")
}

func (sw *StandardWriter) WriteCharData(c *CharData, w *bufio.Writer) {
	if c.IsCData() {
		w.WriteString(`<![CDATA[`)
		w.WriteString(c.Data)
		w.WriteString(`]]>`)
	} else {
		var m EscapeMode
		if sw.Settings.CanonicalText {
			m = EscapeCanonicalText
		} else {
			m = EscapeNormal
		}
		sw.WriteEscape(w, c.Data, m)
	}
}

func (sw *StandardWriter) WriteProcInst(p *ProcInst, w *bufio.Writer) {
	w.WriteString("<?")
	w.WriteString(p.Target)
	if p.Inst != "" {
		w.WriteByte(' ')
		w.WriteString(p.Inst)
	}
	w.WriteString("?>")
}

func (sw *StandardWriter) WriteElement(e *Element, w *bufio.Writer) {
	w.WriteByte('<')
	w.WriteString(e.FullTag())
	for _, a := range e.Attr {
		w.WriteByte(' ')
		sw.WriteAttr(&a, w)
	}
	if len(e.Child) > 0 {
		w.WriteByte('>')
		for _, c := range e.Child {
			sw.WriteToken(c, w)
		}
		w.Write([]byte{'<', '/'})
		w.WriteString(e.FullTag())
		w.WriteByte('>')
	} else {
		if sw.Settings.CanonicalEndTags {
			w.Write([]byte{'>', '<', '/'})
			w.WriteString(e.FullTag())
			w.WriteByte('>')
		} else {
			w.Write([]byte{'/', '>'})
		}
	}
}

func (sw *StandardWriter) WriteAttr(a *Attr, w *bufio.Writer) {
	w.WriteString(a.FullKey())
	w.WriteString(`="`)
	var m EscapeMode
	if sw.Settings.CanonicalAttrVal {
		m = EscapeCanonicalAttr
	} else {
		m = EscapeNormal
	}
	sw.WriteEscape(w, a.Value, m)
	w.WriteByte('"')
}

type StandardEscaper struct{}

func (se *StandardEscaper) WriteEscape(w *bufio.Writer, s string, m EscapeMode) {
	var esc []byte
	last := 0
	for i := 0; i < len(s); {
		r, width := utf8.DecodeRuneInString(s[i:])
		i += width
		switch r {
		case '&':
			esc = []byte("&amp;")
		case '<':
			esc = []byte("&lt;")
		case '>':
			if m == EscapeCanonicalAttr {
				continue
			}
			esc = []byte("&gt;")
		case '\'':
			if m != EscapeNormal {
				continue
			}
			esc = []byte("&apos;")
		case '"':
			if m == EscapeCanonicalText {
				continue
			}
			esc = []byte("&quot;")
		case '\t':
			if m != EscapeCanonicalAttr {
				continue
			}
			esc = []byte("&#x9;")
		case '\n':
			if m != EscapeCanonicalAttr {
				continue
			}
			esc = []byte("&#xA;")
		case '\r':
			if m == EscapeNormal {
				continue
			}
			esc = []byte("&#xD;")
		default:
			if !se.IsInCharacterRange(r) || (r == 0xFFFD && width == 1) {
				esc = []byte("\uFFFD")
				break
			}
			continue
		}
		w.WriteString(s[last : i-width])
		w.Write(esc)
		last = i
	}
	w.WriteString(s[last:])
}

func (*StandardEscaper) IsInCharacterRange(r rune) bool {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}
