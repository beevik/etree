// Package etree provides XML services through an Element Tree abstraction.
package etree

import (
    "bufio"
    "container/list"
    "fmt"
    "io"
)

const sp string = "\n                                                            "

// A Token is an empty interface that represents an Element,
// Comment, CharData or ProcInst.
type Token interface {
    writeTo(w *bufio.Writer)
}

// An Element represents an XML element.  The Children list contains
// Tokens.
type Element struct {
    Name     string
    Attr     []Attr
    Children *list.List
}

// A Comment represents an XML comment
type Comment []byte

// CharData represents character data within XML.
type CharData []byte

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
    Key   string
    Value string
}

// NewElement creates a root-level XML element with the specified name.
func NewElement(name string) *Element {
    return &Element{
        Name:     name,
        Attr:     make([]Attr, 0),
        Children: list.New(),       // list of Tokens
    }
}

// CreateElement creates a child element of the the receiving element and
// gives it the specified name.
func (e *Element) CreateElement(name string) *Element {
    c := NewElement(name)
    e.addChild(c)
    return c
}

// WriteTo serializes the element and its children as XML into
// the writer w.
func (e *Element) WriteTo(w io.Writer) error {
    b := bufio.NewWriter(w)
    e.writeTo(b)
    return b.Flush()
}

// Indent modifies the element tree by inserting CharData entities
// that introduce indentation.  The amount of indenting per depth
// level is equal to spaces.
func (e *Element) Indent(spaces int) {
    e.indent(1, spaces)
}

// indent recursively inserts proper indentation into between an
// XML element's child tokens.
func (e *Element) indent(depth, spaces int) {
    for c := e.Children.Front(); c != nil; {
        n := c.Next()
        e.Children.InsertBefore(indentCharData(depth, spaces), c)
        if ce, ok := c.Value.(*Element); ok {
            ce.indent(depth + 1, spaces)
        }
        c = n
    }
    if b := e.Children.Back(); b != nil {
        e.Children.InsertAfter(indentCharData(depth - 1, spaces), b)
    }
}

// addChild adds a child token to the receiving element.
func (e *Element) addChild(t Token) {
    e.Children.PushBack(t)
}

// writeTo serializes the element to the writer w.
func (e *Element) writeTo(w *bufio.Writer) {
    w.WriteString(fmt.Sprintf("<%s", e.Name))
    for _, a := range e.Attr {
        w.WriteByte(' ')
        a.writeTo(w)
    }
    if e.Children.Len() > 0 {
        w.WriteString(">")
        for c := e.Children.Front(); c != nil; c = c.Next() {
            c.Value.(Token).writeTo(w)
        }
        w.WriteString(fmt.Sprintf("</%s>", e.Name))
    } else {
        w.WriteString("/>")
    }
}

// AddAttr adds an attribute to the element.
func (e *Element) AddAttr(a Attr) {
    e.Attr = append(e.Attr, a)
}

// CreateAttr creates an attribute and adds it to the receiving element.
func (e *Element) CreateAttr(key, value string) Attr {
    a := Attr{key, value}
    e.Attr = append(e.Attr, a)
    return a
}

// writeTo serializes the attribute to the writer.
func (a *Attr) writeTo(w *bufio.Writer) {
    w.WriteString(fmt.Sprintf(`%s="%s"`, a.Key, escape(a.Value)))
}

// newCharData creates an XML character data entity.
func newCharData(charData string) *CharData {
    c := new(CharData)
    *c = CharData(charData)
    return c
}

// CreateCharData creates an XML character data entity and adds it
// as a child of the receiving element.
func (e *Element) CreateCharData(charData string) *CharData {
    c := newCharData(charData)
    e.addChild(c)
    return c
}

// writeTo serializes the character data entity to the writer.
func (c *CharData) writeTo(w *bufio.Writer) {
    w.WriteString(string(*c))
}

// NewComment creates an XML comment.
func newComment(comment string) *Comment {
    c := new(Comment)
    *c = Comment(comment)
    return c
}

// CreateComment creates an XML comment and adds it as a child of the
// receiving element.
func (e *Element) CreateComment(comment string) *Comment {
    c := newComment(comment)
    e.addChild(c)
    return c
}

// writeTo serialies the comment to the writer.
func (c *Comment) writeTo(w *bufio.Writer) {
    w.WriteString(fmt.Sprintf("<!-- %s -->", string(*c)))
}

// indentCharData returns the indentation CharData token for the given
// depth level with the given number of spaces per level.
func indentCharData(depth, spaces int) *CharData {
    c := 1 + depth * spaces
    if c > len(sp) {
        return newCharData(sp)
    } else {
        return newCharData(sp[:c])
    }
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
