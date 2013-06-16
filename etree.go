// Package etree provides XML services through an Element Tree
// abstraction.
package etree

import (
    "bufio"
    "encoding/xml"
    "errors"
    "io"
)

var (
    ErrInvalidFormat = errors.New("etree: invalid XML format")
)

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
type Comment struct {
    Data []byte
}

// CharData represents character data within XML.
type CharData struct {
    Data       []byte
    whitespace bool
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

// ReadFrom reads XML from the reader r and adds the result as
// a new child of the receiving document.
func (d *Document) ReadFrom(r io.Reader) error {
    return d.Element.ReadFrom(r)
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
            newChild[j+1] = newIndentCharData(0, spaces)
        }
        if e, ok := c.(*Element); ok {
            e.indent(1, spaces)
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

// ReadFrom reads XML from the reader r and stores the result as
// a new child of the receiving element.
func (e *Element) ReadFrom(r io.Reader) error {
    stack := []*Element{e}
    dec := xml.NewDecoder(r)
    for {
        t, err := dec.RawToken()
        if err == io.EOF {
            break
        } else if err != nil {
            return err
        } else if len(stack) == 0 {
            return ErrInvalidFormat
        }
        top := stack[len(stack)-1]
        switch t := t.(type) {
        case xml.StartElement:
            e := top.CreateElement(t.Name.Local)
            for _, a := range t.Attr {
                e.CreateAttr(a.Name.Local, a.Value)
            }
            stack = append(stack, e)
        case xml.EndElement:
            stack[len(stack)-1] = nil
            stack = stack[:len(stack)-1]
        case xml.CharData:
            data := copyBytes(t)
            cd := &CharData{Data: data, whitespace: isWhitespace(data)}
            top.Child = append(top.Child, cd)
        case xml.Comment:
            top.Child = append(top.Child, &Comment{Data: copyBytes(t)})
        case xml.ProcInst:
            top.Child = append(
                top.Child,
                &ProcInst{Target: []byte(t.Target), Inst: copyBytes(t.Inst)},
            )
        }
    }
    return nil
}

// WriteTo serializes the element and its children as XML into
// the writer w.
func (e *Element) WriteTo(w io.Writer) error {
    b := bufio.NewWriter(w)
    e.writeTo(b)
    return b.Flush()
}

// ChildElements returns all elements that are children of the
// receiving element.
func (e *Element) ChildElements() []*Element {
    elements := make([]*Element, 0)
    for _, c := range e.Child {
        if e, ok := c.(*Element); ok {
            elements = append(elements, e)
        }
    }
    return elements
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
        if cd, ok := c.(*CharData); ok && cd.whitespace {
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
        if cd, ok := c.(*CharData); ok && cd.whitespace {
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
func newCharData(data string, whitespace bool) *CharData {
    return &CharData{
        Data:       []byte(data),
        whitespace: whitespace,
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
    return &Comment{Data: []byte(comment)}
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
    w.Write(c.Data)
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
    return &CharData{
        Data:       crSpaces(depth * spaces),
        whitespace: true,
    }
}
