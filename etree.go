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

// Package etree provides XML services through an Element Tree
// abstraction.
package etree

import (
	"bufio"
	"encoding/xml"
	"errors"
	"io"
)

const (
	NoIndent = -1 // Use with Indent to turn off indenting
)

var (
	ErrInvalidFormat = errors.New("etree: invalid XML format")
)

// A Token is an empty interface that represents an Element,
// Comment, CharData, or ProcInst.
type Token interface {
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
	Tag    string   // The element tag
	Attr   []Attr   // The element's key-value attribute pairs
	Child  []Token  // The element's child tokens (elements, comments, etc.)
	Parent *Element // The element's parent element
}

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
	Key   string
	Value string
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

// A ProcInst represents an XML processing instruction.
type ProcInst struct {
	Target string
	Inst   string
}

// NewDocument creates an empty XML document and returns it.
func NewDocument() *Document {
	return &Document{Element{Child: make([]Token, 0)}}
}

// ReadFrom reads XML from the reader r and adds the result as
// a new child of the receiving document.
func (d *Document) ReadFrom(r io.Reader) error {
	return d.Element.readFrom(r)
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
// Pass etree.NoIndent for spaces if you want no indentation at all.
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
// gives it the specified name.
func (e *Element) CreateElement(tag string) *Element {
	c := &Element{
		Tag:    tag,
		Attr:   make([]Attr, 0),
		Child:  make([]Token, 0),
		Parent: e,
	}
	e.addChild(c)
	return c
}

// An element stack is a simple stack of elements used by readFrom.
type elementStack []*Element

func (s *elementStack) push(e *Element) {
	*s = append(*s, e)
}

func (s *elementStack) pop() {
	(*s)[len(*s)-1] = nil
	*s = (*s)[:len(*s)-1]
}

func (s *elementStack) peek() *Element {
	return (*s)[len(*s)-1]
}

// ReadFrom reads XML from the reader r and stores the result as
// a new child of the receiving element.
func (e *Element) readFrom(r io.Reader) error {
	stack := elementStack{e}
	dec := xml.NewDecoder(r)
	for {
		t, err := dec.RawToken()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case len(stack) == 0:
			return ErrInvalidFormat
		}

		top := stack.peek()

		switch t := t.(type) {
		case xml.StartElement:
			e := top.CreateElement(t.Name.Local)
			for _, a := range t.Attr {
				e.CreateAttr(a.Name.Local, a.Value)
			}
			stack.push(e)
		case xml.EndElement:
			stack.pop()
		case xml.CharData:
			data := string(t)
			top.createCharData(data, isWhitespace(data))
		case xml.Comment:
			top.CreateComment(string(t))
		case xml.ProcInst:
			top.CreateProcInst(t.Target, string(t.Inst))
		}
	}
}

// SelectAttr finds an element attribute matching the requested key
// and returns it if found.
func (e *Element) SelectAttr(key string) *Attr {
	for _, a := range e.Attr {
		if a.Key == key {
			return &a
		}
	}
	return nil
}

// SelectAttrValue finds an element attribute matching the requested key
// and returns its value if found.  If it is not found, the dflt value
// is returned instead.
func (e *Element) SelectAttrValue(key, dflt string) string {
	for _, a := range e.Attr {
		if a.Key == key {
			return a.Value
		}
	}
	return dflt
}

// ChildElements returns all elements that are children of the
// receiving element.
func (e *Element) ChildElements() []*Element {
	elements := make([]*Element, 0)
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok {
			elements = append(elements, c)
		}
	}
	return elements
}

// SelectElement returns the first child element with the given tag.
func (e *Element) SelectElement(tag string) *Element {
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok && c.Tag == tag {
			return c
		}
	}
	return nil
}

// SelectElements returns a slice of all child elements with the given tag.
func (e *Element) SelectElements(tag string) []*Element {
	elements := make([]*Element, 0)
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok && c.Tag == tag {
			elements = append(elements, c)
		}
	}
	return elements
}

// FindElements finds all elements matching the XPath-like path string and
// returns them in a slice.
func (e *Element) FindElements(path string) []*Element {
	elements := make([]*Element, 0)
	// TODO: Write me
	return elements
}

// Indent modifies the element's element tree by inserting
// CharData entities containing carriage returns and indentation.
// The amount of indenting per depth level is equal to spaces.
// Use etree.NoIndent for spaces if you want no indentation at all.
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

	oldChild := e.Child
	e.Child = make([]Token, 0, n*2+1)
	isCharData := false
	for _, c := range oldChild {
		_, isCharData = c.(*CharData)
		if !isCharData && spaces >= 0 {
			e.addChild(newIndentCharData(depth, spaces))
		}
		e.addChild(c)
		if ce, ok := c.(*Element); ok {
			ce.indent(depth+1, spaces)
		}
	}
	if !isCharData && spaces >= 0 {
		e.addChild(newIndentCharData(depth-1, spaces))
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

// writeTo serializes the element to the writer w.
func (e *Element) writeTo(w *bufio.Writer) {
	w.WriteByte('<')
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
func (e *Element) CreateAttr(key, value string) Attr {
	a := Attr{key, value}
	e.Attr = append(e.Attr, a)
	return a
}

// CreateElementIterator creates a child element iterator that iterates
// through all child elements with the given tag.
func (e *Element) CreateElementIterator(tag string) *ElementIterator {
	i := &ElementIterator{
		parent: e,
		tag:    tag,
		index:  -1,
	}
	i.Next()
	return i
}

// writeTo serializes the attribute to the writer.
func (a *Attr) writeTo(w *bufio.Writer) {
	w.WriteString(a.Key)
	w.WriteString(`="`)
	w.WriteString(a.Value)
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

// writeTo serialies the comment to the writer.
func (c *Comment) writeTo(w *bufio.Writer) {
	w.WriteString("<!--")
	w.WriteString(c.Data)
	w.WriteString("-->")
}

// newProcInst creates a new processing instruction.
func newProcInst(target, inst string) *ProcInst {
	return &ProcInst{Target: target, Inst: inst}
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
	w.WriteString("<?")
	w.WriteString(p.Target)
	w.WriteByte(' ')
	w.WriteString(p.Inst)
	w.WriteString("?>")
}

// newIndentCharData returns the indentation CharData token for the given
// depth level with the given number of spaces per level.
func newIndentCharData(depth, spaces int) *CharData {
	return newCharData(crSpaces(depth*spaces), true)
}

// An ElementIterator allows the caller to iterate through the child elements
// of an element.
type ElementIterator struct {
	Element *Element
	parent  *Element
	tag     string
	index   int
}

// Valid returns true if the ElementIterator currently points to a valid element.
func (i *ElementIterator) Valid() bool {
	return i.Element != nil
}

// Next advances to the next child element with the appropriate tag.
func (i *ElementIterator) Next() {
	for {
		i.index++
		if i.index >= len(i.parent.Child) {
			i.Element = nil
			return
		}
		c := i.parent.Child[i.index]
		if n, ok := c.(*Element); ok && n.Tag == i.tag {
			i.Element = n
			return
		}
	}
}
