// Package etree provides XML services through an Element Tree abstraction.
package etree

import (
    "bufio"
    "fmt"
    "io"
)

const sp string = "                                                     "

// An Element represents an XML element.
type Element struct {
    Name  string
    Attr  []Attr
    Child []*Element
}

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
    Key   string
    Value string
}

// WriteTo serializes the element and its children as XML into
// the writer w.
func (e *Element) WriteTo(w io.Writer) error {
    b := bufio.NewWriter(w)
    e.write(b, 0)
    return b.Flush()
}

// NewElement creates an element with the specified name.
func NewElement(name string) *Element {
    return &Element{
        Name:  name,
        Attr:  make([]Attr, 0),
        Child: make([]*Element, 0),
    }
}

// AddAttr creates an attribute and adds it to the element.
func (e *Element) AddAttr(key, value string) {
    e.Attr = append(e.Attr, Attr{key, value})
}

// AddChild adds a child element to the element.
func (e *Element) AddChild(c *Element) {
    e.Child = append(e.Child, c)
}

// CreateChild creates a child element of the element with the given name.
func (e *Element) CreateChild(name string) *Element {
    c := NewElement(name)
    e.Child = append(e.Child, c)
    return c
}

// write serializes the element to the writer w at the specified indent depth.
func (e *Element) write(w *bufio.Writer, depth int) {
    w.WriteString(fmt.Sprintf("%s<%s", spaces(depth), e.Name))
    for _, a := range e.Attr {
        w.WriteByte(' ')
        a.write(w)
    }
    if len(e.Child) > 0 {
        w.WriteString(">\n")
        for _, c := range e.Child {
            c.write(w, depth+1)
        }
        w.WriteString(fmt.Sprintf("%s</%s>\n", spaces(depth), e.Name))
    } else {
        w.WriteString("/>\n")
    }
}

// write serializes the attribute to the writer.
func (a *Attr) write(w *bufio.Writer) {
    w.WriteString(fmt.Sprintf(`%s="%s"`, a.Key, escape(a.Value)))
}

// spaces outputs a string of spaces for the indent level of depth.
func spaces(depth int) string {
    return sp[0 : depth*2]
}

// escape generates an escaped XML string.
func escape(s string) string {
    buf := make([]byte, 0, len(s))
    for i := 0; i < len(s); i++ {
        switch {
        case s[i] == '&':
            buf = append(buf, []byte{'&', 'a', 'm', 'p', ';'}...)
        case s[i] == '\'':
            buf = append(buf, []byte{'&', 'a', 'p', 'o', 's', ';'}...)
        case s[i] == '<':
            buf = append(buf, []byte{'&', 'l', 't', ';'}...)
        case s[i] == '>':
            buf = append(buf, []byte{'&', 'g', 't', ';'}...)
        default:
            buf = append(buf, s[i])
        }
    }
    return string(buf)
}
