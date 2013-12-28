// Copyright 2013 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"errors"
	"strconv"
	"strings"
)

var (
	errPath = errors.New("etree: invalid path")
)

// A Path is an object that represents an optimized version of an
// XPath-like string.  Although the path strings are XPath-like,
// only the following limited syntax is supported:
//
//     .               Selects the current element
//     ..              Selects the parent of the current element
//     *               Selects all child elements
//     //              Selects all descendants of the current element
//     tag             Selects all child elements with the given tag
//     [#]             Selects the element with the given index (1-based)
//     [@attrib]       Selects all elements with the given attribute
//     [@attrib='val'] Selects all elements with the given attribute set to val
//     [tag]           Selects all elements with a child element named tag
//     [tag='val']     Selects all elements with a cihld element named tag and text equal to val
//
// Examples:
//
// Select the title elements of all book elements with a category attribute
// of WEB:
//     //book[@category='WEB']/title
//
// Select the first book element with a title child containing the text
// 'Great Expectations':
//     .//book[title='Great Expectations'][1]
//
// Select all grandchildren with an attribute 'language' equal to 'english'
// and a parent element tag of 'book':
//     book/*[@language='english']
//
// Select all book elements whose title child has a language of 'french':
//     //book/title[@language='french']/..
type Path struct {
	segments []segment
}

// CompilePath creates an optimized version of an XPath-like string that
// can be used to query elements in an element tree.
func CompilePath(path string) (Path, error) {
	segments, err := parsePath(path)
	if err != nil {
		return Path{nil}, err
	}
	return Path{segments}, nil
}

// MustCompilePath creates an optimized version of an XPath-like string that
// can be used to query elements in an element tree.  Panics if an error
// occurs.  Use this function to create Paths when you know the path is
// valid (i.e., if it's hard-coded).
func MustCompilePath(path string) Path {
	segments, err := parsePath(path)
	if err != nil {
		panic(err)
	}
	return Path{segments}
}

// A segment is a portion of a path between "/" characters.
// It contains one selector and zero or more [filters].
type segment struct {
	sel     selector
	filters []filter
}

func (seg *segment) apply(e *Element, p *pather) {
	seg.sel.apply(e, p)
	for _, f := range seg.filters {
		f.apply(p)
	}
}

// A selector selects XML elements for consideration by the
// path traversal.
type selector interface {
	apply(e *Element, p *pather)
}

// A filter pares down a list of candidate XML elements based
// on a path filter in [brackets].
type filter interface {
	apply(p *pather)
}

// A pather is helper object used to traverse an element tree with
// a Path object.  It collects and deduplicates elements matching
// the path query.
type pather struct {
	queue      fifo
	results    []*Element
	inResults  map[*Element]bool
	candidates []*Element
	scratch    []*Element // used by filters
}

// A node represents an element and the remaining path segments that
// should be applied against it by the pather.
type node struct {
	e        *Element
	segments []segment
}

func newPather() *pather {
	return &pather{
		results:    make([]*Element, 0),
		inResults:  make(map[*Element]bool),
		candidates: make([]*Element, 0),
		scratch:    make([]*Element, 0),
	}
}

// traverse follows the path from the element e, collecting
// and then returning all elements that match the path's selectors
// and filters.
func (p *pather) traverse(e *Element, path Path) []*Element {
	for p.queue.add(node{e, path.segments}); p.queue.len() > 0; {
		p.eval(p.queue.remove().(node))
	}
	return p.results
}

// eval evalutes the current path node by applying the remaining
// path's selector rules against the node's element.
func (p *pather) eval(n node) {
	p.candidates = p.candidates[0:0]
	seg, remain := n.segments[0], n.segments[1:]
	seg.apply(n.e, p)

	if len(remain) == 0 {
		for _, c := range p.candidates {
			if in := p.inResults[c]; !in {
				p.inResults[c] = true
				p.results = append(p.results, c)
			}
		}
	} else {
		for _, c := range p.candidates {
			p.queue.add(node{c, remain})
		}
	}
}

// parsePath parses an XPath-like string describing a path
// through an element tree and returns a slice of segment
// descriptors.
func parsePath(path string) ([]segment, error) {
	// If path starts or ends with //, fix it
	if strings.HasPrefix(path, "//") {
		path = "." + path
	}
	if strings.HasSuffix(path, "//") {
		path = path + "*"
	}

	// Paths cannot be absolute
	if strings.HasPrefix(path, "/") {
		return nil, errPath
	}

	// Split path into segment objects
	segments := make([]segment, 0)
	for _, s := range strings.Split(path, "/") {
		segments = append(segments, parseSegment(s))
	}
	return segments, nil
}

// parseSegment parses a path segment between / characters.
func parseSegment(path string) segment {
	pieces := strings.Split(path, "[")
	seg := segment{
		sel:     parseSelector(pieces[0]),
		filters: make([]filter, 0),
	}
	for i := 1; i < len(pieces); i++ {
		fpath := pieces[i]
		if fpath[len(fpath)-1] != ']' {
			panic(errPath)
		}
		seg.filters = append(seg.filters, parseFilter(fpath[:len(fpath)-1]))
	}
	return seg
}

// parseSelector parses a selector at the start of a path segment.
func parseSelector(path string) selector {
	switch path {
	case ".":
		return new(selectSelf)
	case "..":
		return new(selectParent)
	case "*":
		return new(selectChildren)
	case "":
		return new(selectDescendants)
	default:
		return newSelectChildrenTag(path)
	}
}

// parseFilter parses a path filter contained within [brackets].
func parseFilter(path string) filter {
	if len(path) == 0 {
		panic(errPath)
	}

	// Filter contains [@attr='val'] or [tag='val']?
	eqindex := strings.Index(path, "='")
	if eqindex >= 0 {
		rindex := nextIndex(path, "'", eqindex+2)
		if rindex != len(path)-1 {
			panic(errPath)
		}
		switch {
		case path[0] == '@':
			return newFilterAttrVal(path[1:eqindex], path[eqindex+2:rindex])
		default:
			return newFilterChildText(path[:eqindex], path[eqindex+2:rindex])
		}
	}

	// Filter contains [@attr], [N] or [tag]
	switch {
	case path[0] == '@':
		return newFilterAttr(path[1:])
	case isInteger(path):
		pos, _ := strconv.Atoi(path)
		if pos == 0 {
			pos = 1 // force to 1-based
		}
		return newFilterPos(pos - 1)
	default:
		return newFilterChild(path)
	}
}

// selectSelf selects the current element into the candidate list.
type selectSelf struct{}

func (s *selectSelf) apply(e *Element, p *pather) {
	p.candidates = append(p.candidates, e)
}

// selectParent selects the element's parent into the candidate list.
type selectParent struct{}

func (s *selectParent) apply(e *Element, p *pather) {
	if e.Parent != nil {
		p.candidates = append(p.candidates, e.Parent)
	}
}

// selectChildren selects the element's child elements into the
// candidate list.
type selectChildren struct{}

func (s *selectChildren) apply(e *Element, p *pather) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok {
			p.candidates = append(p.candidates, c)
		}
	}
}

// selectDescendants selects all descendant child elements
// of the element into the candidate list.
type selectDescendants struct{}

func (s *selectDescendants) apply(e *Element, p *pather) {
	queue := fifo{}
	for queue.add(e); queue.len() > 0; {
		e := queue.remove().(*Element)
		p.candidates = append(p.candidates, e)
		for _, c := range e.Child {
			if c, ok := c.(*Element); ok {
				queue.add(c)
			}
		}
	}
}

// Break a string at ':' and return the two parts.
func decompose(str string) (space, key string) {
	colon := strings.IndexByte(str, ':')
	if colon == -1 {
		return "", str
	}
	return str[:colon], str[colon+1:]
}

// selectChildrenTag selects into the candidate list all child
// elements of the element having the specified tag.
type selectChildrenTag struct {
	space, tag string
}

func newSelectChildrenTag(path string) *selectChildrenTag {
	s, l := decompose(path)
	return &selectChildrenTag{s, l}
}

func (s *selectChildrenTag) apply(e *Element, p *pather) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && c.Space == s.space && c.Tag == s.tag {
			p.candidates = append(p.candidates, c)
		}
	}
}

// filterPos filters the candidate list, keeping only the
// candidate at the specified index.
type filterPos struct {
	index int
}

func newFilterPos(pos int) *filterPos {
	return &filterPos{pos}
}

func (f *filterPos) apply(p *pather) {
	p.scratch = p.scratch[:0]
	if f.index < len(p.candidates) {
		p.scratch = append(p.scratch, p.candidates[f.index])
	}
	p.candidates = p.scratch
}

// filterAttr filters the candidate list for elements having
// the specified attribute.
type filterAttr struct {
	space, key string
}

func newFilterAttr(str string) *filterAttr {
	s, l := decompose(str)
	return &filterAttr{s, l}
}

func (f *filterAttr) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		if a := c.SelectAttrFull(f.space, f.key); a != nil {
			p.scratch = append(p.scratch, c)
		}
	}
	p.candidates = p.scratch
}

// filterAttrVal filters the candidate list for elements having
// the specified attribute with the specified value.
type filterAttrVal struct {
	space, key, val string
}

func newFilterAttrVal(str, value string) *filterAttrVal {
	s, l := decompose(str)
	return &filterAttrVal{s, l, value}
}

func (f *filterAttrVal) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		if a := c.SelectAttrFull(f.space, f.key); a != nil && a.Value == f.val {
			p.scratch = append(p.scratch, c)
		}
	}
	p.candidates = p.scratch
}

// filterChild filters the candidate list for elements having
// a child element with the specified tag.
type filterChild struct {
	space, tag string
}

func newFilterChild(str string) *filterChild {
	s, l := decompose(str)
	return &filterChild{s, l}
}

func (f *filterChild) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok &&
				cc.Space == f.space &&
				cc.Tag == f.tag {
				p.scratch = append(p.scratch, c)
			}
		}
	}
	p.candidates = p.scratch
}

// filterChildText filters the candidate list for elements having
// a child element with the specified tag and text.
type filterChildText struct {
	space, tag, text string
}

func newFilterChildText(str, text string) *filterChildText {
	s, l := decompose(str)
	return &filterChildText{s, l, text}
}

func (f *filterChildText) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok &&
				cc.Space == f.space &&
				cc.Tag == f.tag &&
				cc.Text() == f.text {
				p.scratch = append(p.scratch, c)
			}
		}
	}
	p.candidates = p.scratch
}
