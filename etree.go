// Copyright 2013 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package etree provides XML services through an Element Tree
// abstraction.
package etree

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	// NoIndent is used with Indent to disable all indenting.
	NoIndent = -1
)

var (
	// ErrXML is returned when XML parsing fails due to incorrect formatting.
	ErrXML = errors.New("etree: invalid XML format")

	// ErrPath is returned when an invalid etree path is provided.
	ErrPath = errors.New("etree: invalid path")
)

// A Token is an empty interface that represents an Element,
// Comment, CharData, or ProcInst.
type Token interface {
	dup(parent *Element) Token
	writeTo(w *bufio.Writer)
}

// A Document is the root level object in an etree.  It represents the
// XML document as a whole.  It embeds an Element type but only uses the
// its Child tokens.
type Document struct {
	Element
}

// An Element represents an XML element, its attributes, and its child tokens.
type Element struct {
	Space, Tag string   // The element's namespace and tag
	Attr       []Attr   // The element's key-value attribute pairs
	Child      []Token  // The element's child tokens (elements, comments, etc.)
	Parent     *Element // The element's parent element
}

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
	Space, Key string // The attribute's namespace and key
	Value      string // The attribute value string
}

// A Comment represents an XML comment.
type Comment struct {
	Data string
}

// CharData represents character data within XML.
type CharData struct {
	Data       string
	whitespace bool
}

// A Directive represents an XML directive.
type Directive struct {
	Data string
}

// A ProcInst represents an XML processing instruction.
type ProcInst struct {
	Target string
	Inst   string
}

// CreateDocument creates a new XML document with root
// as the root element.
func CreateDocument(root *Element) *Document {
	doc := NewDocument()
	doc.Child = append(doc.Child, root)
	return doc
}

// NewDocument creates an empty XML document and returns it.
func NewDocument() *Document {
	return &Document{Element{Child: make([]Token, 0)}}
}

// Copy returns a recursive, deep copy of the document.
func (d *Document) Copy() *Document {
	return &Document{*(d.dup(nil).(*Element))}
}

// Root returns the root element of the document, or nil if there is no root
// element.
func (d *Document) Root() *Element {
	for _, t := range d.Child {
		if c, ok := t.(*Element); ok {
			return c
		}
	}
	return nil
}

// ReadFrom reads XML from the reader r into the document d.
// It returns the number of bytes read and any error encountered.
func (d *Document) ReadFrom(r io.Reader) (n int64, err error) {
	return d.Element.readFrom(r)
}

// ReadFromFile reads XML from the string s into the document d.
func (d *Document) ReadFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = d.ReadFrom(f)
	return err
}

// ReadFromBytes reads XML from the byte slice b into the document d.
func (d *Document) ReadFromBytes(b []byte) error {
	_, err := d.ReadFrom(bytes.NewReader(b))
	return err
}

// ReadFromString reads XML from the string s into the document d.
func (d *Document) ReadFromString(s string) error {
	_, err := d.ReadFrom(strings.NewReader(s))
	return err
}

// WriteTo serializes an XML document into the writer w. It
// returns the number of bytes written and any error encountered.
func (d *Document) WriteTo(w io.Writer) (n int64, err error) {
	cw := newCountWriter(w)
	b := bufio.NewWriter(cw)
	for _, c := range d.Child {
		c.writeTo(b)
	}
	err, n = b.Flush(), cw.bytes
	return
}

// WriteToFile serializes an XML document into the file named
// filename.
func (d *Document) WriteToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = d.WriteTo(f)
	return err
}

// WriteToBytes serializes the XML document into a slice of
// bytes.
func (d *Document) WriteToBytes() (b []byte, err error) {
	var buf bytes.Buffer
	if _, err = d.WriteTo(&buf); err != nil {
		return
	}
	return buf.Bytes(), nil
}

// WriteToString serializes the XML document into a string.
func (d *Document) WriteToString() (s string, err error) {
	var b []byte
	if b, err = d.WriteToBytes(); err != nil {
		return
	}
	return string(b), nil
}

type indentFunc func(depth int) string

// Indent modifies the document's element tree by inserting
// CharData entities containing carriage returns and indentation.
// The amount of indentation per depth level is given as spaces.
// Pass etree.NoIndent for spaces if you want no indentation at all.
func (d *Document) Indent(spaces int) {
	var indent indentFunc
	switch {
	case spaces < 0:
		indent = func(depth int) string { return "" }
	default:
		indent = func(depth int) string { return crIndent(depth*spaces, crsp) }
	}
	d.Element.indent(0, indent)
}

// IndentTabs modifies the document's element tree by inserting
// CharData entities containing carriage returns and tabs for
// indentation.  One tab is used per indentation level.
func (d *Document) IndentTabs() {
	indent := func(depth int) string { return crIndent(depth, crtab) }
	d.Element.indent(0, indent)
}

// Copy creates a parentless, recursive, deep copy of the element and
// all its attributes and children. The returned element has no
// parent but can be parented to a document using the CreateDocument
// function.
func (e *Element) Copy() *Element {
	var parent *Element
	return e.dup(parent).(*Element)
}

// Text returns the characters immediately following the element's
// opening tag.
func (e *Element) Text() string {
	if len(e.Child) == 0 {
		return ""
	}
	if cd, ok := e.Child[0].(*CharData); ok {
		return cd.Data
	}
	return ""
}

// SetText replaces an element's subsidiary CharData text with a new
// string.
func (e *Element) SetText(text string) {
	if len(e.Child) > 0 {
		if cd, ok := e.Child[0].(*CharData); ok {
			cd.Data = text
			return
		}
	}
	e.Child = append(e.Child, nil)
	copy(e.Child[1:], e.Child[0:])
	e.Child[0] = newCharData(text, false)
}

// CreateElement creates a child element of the receiving element and
// gives it the specified tag.  The tag may be prefixed by a namespace
// and a colon.
func (e *Element) CreateElement(tag string) *Element {
	space, stag := spaceDecompose(tag)
	return e.createElement(space, stag)
}

// createElement is a helper function that creates new elements.
func (e *Element) createElement(space, tag string) *Element {
	c := &Element{
		Space:  space,
		Tag:    tag,
		Attr:   make([]Attr, 0),
		Child:  make([]Token, 0),
		Parent: e,
	}
	e.addChild(c)
	return c
}

// RemoveElement removes the given child element. If an identical child element
// does not exist, nil is returned.
func (e *Element) RemoveElement(el *Element) *Element {
	for i, t := range e.Child {
		if c, ok := t.(*Element); ok && c == el {
			e.Child = append(e.Child[0:i], e.Child[i+1:]...)
			c.Parent = nil
			return c
		}
	}
	return nil
}

// ReadFrom reads XML from the reader r and stores the result as
// a new child of the receiving element.
func (e *Element) readFrom(ri io.Reader) (n int64, err error) {
	r := newCountReader(ri)
	dec := xml.NewDecoder(r)
	var stack stack
	stack.push(e)
	for {
		t, err := dec.RawToken()
		switch {
		case err == io.EOF:
			return r.bytes, nil
		case err != nil:
			return r.bytes, err
		case stack.empty():
			return r.bytes, ErrXML
		}

		top := stack.peek().(*Element)

		switch t := t.(type) {
		case xml.StartElement:
			e := top.createElement(t.Name.Space, t.Name.Local)
			for _, a := range t.Attr {
				e.createAttr(a.Name.Space, a.Name.Local, a.Value)
			}
			stack.push(e)
		case xml.EndElement:
			stack.pop()
		case xml.CharData:
			data := string(t)
			top.createCharData(data, isWhitespace(data))
		case xml.Comment:
			top.CreateComment(string(t))
		case xml.Directive:
			top.CreateDirective(string(t))
		case xml.ProcInst:
			top.CreateProcInst(t.Target, string(t.Inst))
		}
	}
}

// SelectAttr finds an element attribute matching the requested key
// and returns it if found. The key may be prefixed by a namespace
// and a colon.
func (e *Element) SelectAttr(key string) *Attr {
	space, skey := spaceDecompose(key)
	for i, a := range e.Attr {
		if spaceMatch(space, a.Space) && skey == a.Key {
			return &e.Attr[i]
		}
	}
	return nil
}

// SelectAttrValue finds an element attribute matching the requested key
// and returns its value if found. The key may be prefixed by a namespace
// and a colon. If the key is not found, the dflt value is returned instead.
func (e *Element) SelectAttrValue(key, dflt string) string {
	space, skey := spaceDecompose(key)
	for _, a := range e.Attr {
		if spaceMatch(space, a.Space) && skey == a.Key {
			return a.Value
		}
	}
	return dflt
}

// ChildElements returns all elements that are children of the
// receiving element.
func (e *Element) ChildElements() []*Element {
	var elements []*Element
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok {
			elements = append(elements, c)
		}
	}
	return elements
}

// SelectElement returns the first child element with the given tag.
// The tag may be prefixed by a namespace and a colon.
func (e *Element) SelectElement(tag string) *Element {
	space, stag := spaceDecompose(tag)
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok && spaceMatch(space, c.Space) && stag == c.Tag {
			return c
		}
	}
	return nil
}

// SelectElements returns a slice of all child elements with the given tag.
// The tag may be prefixed by a namespace and a colon.
func (e *Element) SelectElements(tag string) []*Element {
	space, stag := spaceDecompose(tag)
	var elements []*Element
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok && spaceMatch(space, c.Space) && stag == c.Tag {
			elements = append(elements, c)
		}
	}
	return elements
}

// FindElement returns the first element matched by the XPath-like
// path string. Panics if an invalid path string is supplied.
func (e *Element) FindElement(path string) *Element {
	return e.FindElementPath(MustCompilePath(path))
}

// FindElementPath returns the first element matched by the XPath-like
// path string.
func (e *Element) FindElementPath(path Path) *Element {
	p := newPather()
	elements := p.traverse(e, path)
	switch {
	case len(elements) > 0:
		return elements[0]
	default:
		return nil
	}
}

// FindElements returns a slice of elements matched by the XPath-like
// path string. Panics if an invalid path string is supplied.
func (e *Element) FindElements(path string) []*Element {
	return e.FindElementsPath(MustCompilePath(path))
}

// FindElementsPath returns a slice of elements matched by the Path object.
func (e *Element) FindElementsPath(path Path) []*Element {
	p := newPather()
	return p.traverse(e, path)
}

// indent recursively inserts proper indentation between an
// XML element's child tokens.
func (e *Element) indent(depth int, indent indentFunc) {
	e.stripIndent()
	n := len(e.Child)
	if n == 0 {
		return
	}

	oldChild := e.Child
	e.Child = make([]Token, 0, n*2+1)
	isCharData := false
	for i, c := range oldChild {
		_, isCharData = c.(*CharData)
		if !isCharData && !(i == 0 && depth == 0) {
			e.addChild(newCharData(indent(depth), true))
		}
		e.addChild(c)
		if ce, ok := c.(*Element); ok {
			ce.indent(depth+1, indent)
		}
	}
	if !isCharData {
		e.addChild(newCharData(indent(depth-1), true))
	}
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

// dup duplicates the element.
func (e *Element) dup(parent *Element) Token {
	ne := &Element{
		Space:  e.Space,
		Tag:    e.Tag,
		Attr:   make([]Attr, len(e.Attr)),
		Child:  make([]Token, len(e.Child)),
		Parent: parent,
	}
	for i, t := range e.Child {
		ne.Child[i] = t.dup(ne)
	}
	for i, a := range e.Attr {
		ne.Attr[i] = a
	}
	return ne
}

// writeTo serializes the element to the writer w.
func (e *Element) writeTo(w *bufio.Writer) {
	w.WriteByte('<')
	if e.Space != "" {
		w.WriteString(e.Space)
		w.WriteByte(':')
	}
	w.WriteString(e.Tag)
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
		if e.Space != "" {
			w.WriteString(e.Space)
			w.WriteByte(':')
		}
		w.WriteString(e.Tag)
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
// The key may be prefixed by a namespace and a colon. If an attribute with
// the key already exists, its value is replaced.
func (e *Element) CreateAttr(key, value string) *Attr {
	space, skey := spaceDecompose(key)
	return e.createAttr(space, skey, value)
}

// createAttr is a helper function that creates attributes.
func (e *Element) createAttr(space, key, value string) *Attr {
	for i, a := range e.Attr {
		if space == a.Space && key == a.Key {
			e.Attr[i].Value = value
			return &e.Attr[i]
		}
	}
	a := Attr{space, key, value}
	e.Attr = append(e.Attr, a)
	return &e.Attr[len(e.Attr)-1]
}

// RemoveAttr removes and returns the first attribute of the element whose key
// matches the given key. The key may be prefixed by a namespace and a colon.
// If an equal attribute does not exist, nil is returned.
func (e *Element) RemoveAttr(key string) *Attr {
	space, skey := spaceDecompose(key)
	for i, a := range e.Attr {
		if space == a.Space && skey == a.Key {
			e.Attr = append(e.Attr[0:i], e.Attr[i+1:]...)
			return &a
		}
	}
	return nil
}

// writeTo serializes the attribute to the writer.
func (a *Attr) writeTo(w *bufio.Writer) {
	if a.Space != "" {
		w.WriteString(a.Space)
		w.WriteByte(':')
	}
	w.WriteString(a.Key)
	w.WriteString(`="`)
	w.WriteString(escape(a.Value))
	w.WriteByte('"')
}

// newCharData creates an XML character data entity.
func newCharData(data string, whitespace bool) *CharData {
	return &CharData{Data: data, whitespace: whitespace}
}

// CreateCharData creates an XML character data entity and adds it
// as a child of the receiving element.
func (e *Element) CreateCharData(data string) *CharData {
	return e.createCharData(data, false)
}

// CreateCharData creates an XML character data entity and adds it
// as a child of the receiving element.
func (e *Element) createCharData(data string, whitespace bool) *CharData {
	c := newCharData(data, whitespace)
	e.addChild(c)
	return c
}

// dup duplicates the character data.
func (c *CharData) dup(parent *Element) Token {
	return newCharData(c.Data, c.whitespace)
}

// writeTo serializes the character data entity to the writer.
func (c *CharData) writeTo(w *bufio.Writer) {
	w.WriteString(escape(c.Data))
}

// NewComment creates an XML comment.
func newComment(comment string) *Comment {
	return &Comment{Data: comment}
}

// CreateComment creates an XML comment and adds it as a child of the
// receiving element.
func (e *Element) CreateComment(comment string) *Comment {
	c := newComment(comment)
	e.addChild(c)
	return c
}

// dup duplicates the comment.
func (c *Comment) dup(parent *Element) Token {
	return newComment(c.Data)
}

// writeTo serialies the comment to the writer.
func (c *Comment) writeTo(w *bufio.Writer) {
	w.WriteString("<!--")
	w.WriteString(c.Data)
	w.WriteString("-->")
}

// newDirective creates a new XML directive.
func newDirective(data string) *Directive {
	return &Directive{Data: data}
}

// CreateDirective creates an XML directive and adds it as a
// child of the receiving element.
func (e *Element) CreateDirective(data string) *Directive {
	d := newDirective(data)
	e.addChild(d)
	return d
}

// dup duplicates the directive.
func (d *Directive) dup(parent *Element) Token {
	return newDirective(d.Data)
}

// writeTo serializes the XML directive to the writer.
func (d *Directive) writeTo(w *bufio.Writer) {
	w.WriteString("<!")
	w.WriteString(d.Data)
	w.WriteString(">")
}

// newProcInst creates a new processing instruction.
func newProcInst(target, inst string) *ProcInst {
	return &ProcInst{Target: target, Inst: inst}
}

// CreateProcInst creates a processing instruction and adds it as a
// child of the receiving element.
func (e *Element) CreateProcInst(target, inst string) *ProcInst {
	p := newProcInst(target, inst)
	e.addChild(p)
	return p
}

// dup duplicates the procinst.
func (p *ProcInst) dup(parent *Element) Token {
	return newProcInst(p.Target, p.Inst)
}

// writeTo serializes the processing instruction to the writer.
func (p *ProcInst) writeTo(w *bufio.Writer) {
	w.WriteString("<?")
	w.WriteString(p.Target)
	w.WriteByte(' ')
	w.WriteString(p.Inst)
	w.WriteString("?>")
}
