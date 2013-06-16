// Package etree provides XML services through an Element Tree abstraction.
package etree

import (
    "bufio"
    "container/list"
    "io"
)

const sp string = "\n                                                            "

// A Token is an empty interface that represents an Element,
// Comment, CharData or ProcInst.
type Token interface {
    writeTo(w *bufio.Writer)
}

// A Document represents an XML document.  It is essentially
// an element without a name or attributes, and it is never
// serialized directly to XML; only its children are.
type Document struct {
    Element
}

// An Element represents an XML element.  The Children list contains
// Tokens.
type Element struct {
    Name     []byte
    Attr     []Attr
    Children *list.List
}

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
    Key   []byte
    Value []byte
}

// A Comment represents an XML comment
type Comment []byte

// CharData represents character data within XML.
type CharData []byte

// A ProcInst represents an XML processing instruction.
type ProcInst struct {
    Target []byte
    Inst   []byte
}

// NewDocument creates an empty XML document and returns it.
func NewDocument() *Document {
    d := new(Document)
    d.Children = list.New()
    return d
}

// WriteTo serializes an XML document into the writer w.
func (d *Document) WriteTo(w io.Writer) error {
    b := bufio.NewWriter(w)
    for c := d.Children.Front(); c != nil; c = c.Next() {
        c.Value.(Token).writeTo(b)
    }
    return b.Flush()
}

// Indent modifies the element tree by inserting CharData entities
// that introduce indentation.  The amount of indenting per depth
// level is equal to spaces.
func (d *Document) Indent(spaces int) {
    d.indent(0, spaces)
}

// NewElement creates a root-level XML element with the specified name.
func NewElement(name string) *Element {
    return &Element{
        Name:     []byte(name),
        Attr:     make([]Attr, 0),
        Children: list.New(), // list of Tokens
    }
}

// CreateElement creates a child element of the receiving element and
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

// indent recursively inserts proper indentation between an
// XML element's child tokens.
func (e *Element) indent(depth, spaces int) {
    for c := e.Children.Front(); c != nil; {
        n := c.Next()
        if depth > 0 || c != e.Children.Front() {
            e.Children.InsertBefore(indentCharData(depth, spaces), c)
        }
        if ce, ok := c.Value.(*Element); ok {
            ce.indent(depth+1, spaces)
        }
        c = n
    }
    if b := e.Children.Back(); depth > 0 && b != nil {
        e.Children.InsertAfter(indentCharData(depth-1, spaces), b)
    }
}

// addChild adds a child token to the receiving element.
func (e *Element) addChild(t Token) {
    e.Children.PushBack(t)
}

// writeTo serializes the element to the writer w.
func (e *Element) writeTo(w *bufio.Writer) {
    w.WriteByte('<')
    w.Write(e.Name)
    for _, a := range e.Attr {
        w.WriteByte(' ')
        a.writeTo(w)
    }
    if e.Children.Len() > 0 {
        w.WriteString(">")
        for c := e.Children.Front(); c != nil; c = c.Next() {
            c.Value.(Token).writeTo(w)
        }
        w.Write([]byte{'<', '/'})
        w.Write(e.Name)
        w.WriteByte('>')
    } else {
        w.Write([]byte{'/', '>'})
    }
}

// CreateAttr creates an attribute and adds it to the receiving element.
func (e *Element) CreateAttr(key, value string) Attr {
    a := Attr{[]byte(key), []byte(value)}
    e.Attr = append(e.Attr, a)
    return a
}

// writeTo serializes the attribute to the writer.
func (a *Attr) writeTo(w *bufio.Writer) {
    w.Write(a.Key)
    w.Write([]byte{'=', '"'})
    w.Write(a.Value)
    w.WriteByte('"')
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
    w.Write(escape(*c))
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
    w.Write([]byte{'<', '!', '-', '-', ' '})
    w.Write(*c)
    w.Write([]byte{' ', '-', '-', '>'})
}

// newProcInst creates a new processing instruction.
func newProcInst(target, inst string) *ProcInst {
    return &ProcInst{
        Target: []byte(target),
        Inst:   []byte(inst),
    }
}

// CreateProcInst creates a processing instruction and adds it as a
// child of the receiving element
func (e *Element) CreateProcInst(target, inst string) *ProcInst {
    p := newProcInst(target, inst)
    e.addChild(p)
    return p
}

// writeTo serializes the processing instruction to the writer.
func (p *ProcInst) writeTo(w *bufio.Writer) {
    w.Write([]byte{'<', '?'})
    w.Write(p.Target)
    w.WriteByte(' ')
    w.Write(p.Inst)
    w.Write([]byte{'?', '>'})
}

// indentCharData returns the indentation CharData token for the given
// depth level with the given number of spaces per level.
func indentCharData(depth, spaces int) *CharData {
    c := 1 + depth*spaces
    if c > len(sp) {
        return newCharData(sp)
    } else {
        return newCharData(sp[:c])
    }
}

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
