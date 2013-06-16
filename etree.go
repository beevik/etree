// Package etree provides XML services through an Element Tree
// abstraction.
package etree

import (
    "bufio"
    "io"
)

const sp string = "\n                                                            "

// A Token is an empty interface that represents an Element,
// Comment, CharData, Directive or ProcInst.
type Token interface {
    writeTo(w *bufio.Writer)
}

// A Document is the root level object in an etree.  It represents the
// XML document as a whole.  It embeds an Element type but only uses the
// Element type's Child tokens.
type Document struct {
    Element
}

// An Element represents an XML element, its attributes, and its child tokens.
type Element struct {
    Name  []byte
    Attr  []Attr
    Child []Token
}

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
    Key   []byte
    Value []byte
}

// A Comment represents an XML comment.
type Comment []byte

// CharData represents character data within XML.
type CharData struct {
    Data   []byte
    indent bool
}

// A ProcInst represents an XML processing instruction.
type ProcInst struct {
    Target []byte
    Inst   []byte
}

// NewDocument creates an empty XML document and returns it.
func NewDocument() *Document {
    d := new(Document)
    d.Child = make([]Token, 0)
    return d
}

// WriteTo serializes an XML document into the writer w.
func (d *Document) WriteTo(w io.Writer) error {
    b := bufio.NewWriter(w)
    for _, c := range d.Child {
        c.writeTo(b)
    }
    return b.Flush()
}

// Indent modifies the document's element tree by inserting
// CharData entities containing carriage returns and indentation.
// The amount of indenting per depth level is equal to spaces.
func (d *Document) Indent(spaces int) {
    d.indent(0, spaces)
}

// indent recursively inserts indentation CharData entities
// between an XML document's child tokens.
func (d *Document) indent(depth, spaces int) {
    d.stripIndent()
    n := len(d.Child)
    if n == 0 {
        return
    }
    newChild := make([]Token, n*2-1)
    for i, c := range d.Child {
        j := i * 2
        newChild[j] = c
        if j+1 < len(newChild) {
            newChild[j+1] = newIndentCharData(depth, spaces)
        }
        if e, ok := c.(*Element); ok {
            e.indent(depth+1, spaces)
        }
    }
    d.Child = newChild
}

// NewElement creates an XML element with the specified name.
// In most cases, you should use NewDocument and create elements
// with the CreateElement function.
func NewElement(name string) *Element {
    return &Element{
        Name:  []byte(name),
        Attr:  make([]Attr, 0),
        Child: make([]Token, 0),
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

// Indent modifies the element's element tree by inserting
// CharData entities containing carriage returns and indentation.
// The amount of indenting per depth level is equal to spaces.
func (e *Element) Indent(spaces int) {
    e.indent(1, spaces)
}

// indent recursively inserts proper indentation between an
// XML element's child tokens.
func (e *Element) indent(depth, spaces int) {
    e.stripIndent()
    n := len(e.Child)
    if n == 0 {
        return
    }
    newChild := make([]Token, n*2+1)
    for i, c := range e.Child {
        j := i * 2
        newChild[j] = newIndentCharData(depth, spaces)
        newChild[j+1] = c
        if e, ok := c.(*Element); ok {
            e.indent(depth+1, spaces)
        }
    }
    newChild[n*2] = newIndentCharData(depth-1, spaces)
    e.Child = newChild
}

// stripIndent removes any previously inserted indentation.
func (e *Element) stripIndent() {
    // Count the number of non-indent child tokens
    n := len(e.Child)
    for _, c := range e.Child {
        if cd, ok := c.(*CharData); ok && cd.indent {
            n--
        }
    }
    if n == len(e.Child) {
        return
    }

    // Strip out indent CharData
    newChild := make([]Token, n)
    j := 0
    for _, c := range e.Child {
        if cd, ok := c.(*CharData); ok && cd.indent {
            continue
        }
        newChild[j] = c
        j++
    }
    e.Child = newChild
}

// writeTo serializes the element to the writer w.
func (e *Element) writeTo(w *bufio.Writer) {
    w.WriteByte('<')
    w.Write(e.Name)
    for _, a := range e.Attr {
        w.WriteByte(' ')
        a.writeTo(w)
    }
    if len(e.Child) > 0 {
        w.WriteString(">")
        for _, c := range e.Child {
            c.writeTo(w)
        }
        w.Write([]byte{'<', '/'})
        w.Write(e.Name)
        w.WriteByte('>')
    } else {
        w.Write([]byte{'/', '>'})
    }
}

// addChild adds a child token to the receiving element.
func (e *Element) addChild(t Token) {
    e.Child = append(e.Child, t)
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
func newCharData(data string, indent bool) *CharData {
    return &CharData{
        Data:   []byte(data),
        indent: indent,
    }
}

// CreateCharData creates an XML character data entity and adds it
// as a child of the receiving element.
func (e *Element) CreateCharData(data string) *CharData {
    c := newCharData(data, false)
    e.addChild(c)
    return c
}

// writeTo serializes the character data entity to the writer.
func (c *CharData) writeTo(w *bufio.Writer) {
    w.Write(escape(c.Data))
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

// newIndentCharData returns the indentation CharData token for the given
// depth level with the given number of spaces per level.
func newIndentCharData(depth, spaces int) *CharData {
    c := 1 + depth*spaces
    if c > len(sp) {
        return newCharData(sp, true)
    } else {
        return newCharData(sp[:c], true)
    }
}

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
