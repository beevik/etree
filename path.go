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

// NewPath creates an optimized version of an XPath-like string that
// can be used to query elements in an element tree.  Panics if an
// invalid path string is supplied.
func NewPath(path string) Path {
	return Path{parsePath(path)}
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
	queue      []node
	qindex     int
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
		queue:      make([]node, 0),
		results:    make([]*Element, 0),
		inResults:  make(map[*Element]bool),
		candidates: make([]*Element, 0),
		scratch:    make([]*Element, 0),
	}
}

func (p *pather) add(n node) {
	p.queue = append(p.queue, n)
}

func (p *pather) remove() node {
	n := p.queue[p.qindex]
	p.qindex++
	if p.qindex > len(p.queue)/2 && p.qindex > 32 {
		p.rebalance()
	}
	return n
}

func (p *pather) rebalance() {
	count := len(p.queue) - p.qindex
	copy(p.queue[:count], p.queue[p.qindex:])
	p.queue, p.qindex = p.queue[:count], 0
}

func (p *pather) empty() bool {
	return len(p.queue) == p.qindex
}

// traverse follows the path from the element e, collecting
// and then returning all elements that match the path's selectors
// and filters.
func (p *pather) traverse(e *Element, path Path) []*Element {
	for p.add(node{e, path.segments}); !p.empty(); {
		p.eval(p.remove())
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
			p.add(node{c, remain})
		}
	}
}

// parsePath parses an XPath-like string describing a path
// through an element tree and returns a slice of segment
// descriptors.
func parsePath(path string) []segment {
	// If path starts or ends with //, fix it
	if strings.HasPrefix(path, "//") {
		path = "." + path
	}
	if strings.HasSuffix(path, "//") {
		path = path + "*"
	}

	// Paths cannot be absolute
	if strings.HasPrefix(path, "/") {
		panic(errPath)
	}

	// Split path into segment objects
	segments := make([]segment, 0)
	for _, s := range strings.Split(path, "/") {
		segments = append(segments, parseSegment(s))
	}
	return segments
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
		return &selectChildrenTag{path}
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
			return &filterAttrVal{path[1:eqindex], path[eqindex+2 : rindex]}
		default:
			return &filterChildText{path[:eqindex], path[eqindex+2 : rindex]}
		}
	}

	// Filter contains [@attr], [N] or [tag]
	switch {
	case path[0] == '@':
		return &filterAttr{path[1:]}
	case isInteger(path):
		pos, _ := strconv.Atoi(path)
		if pos == 0 {
			pos = 1 // force to 1-based
		}
		return &filterPos{pos - 1}
	default:
		return &filterChild{path}
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
	stack := elementStack{e}
	for !stack.empty() {
		e := stack.pop()
		p.candidates = append(p.candidates, e)
		for _, c := range e.Child {
			if c, ok := c.(*Element); ok {
				stack.push(c)
			}
		}
	}
}

// selectChildrenTag selects into the candidate list all child
// elements of the element having the specified tag.
type selectChildrenTag struct {
	tag string
}

func (s *selectChildrenTag) apply(e *Element, p *pather) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && c.Tag == s.tag {
			p.candidates = append(p.candidates, c)
		}
	}
}

// filterPos filters the candidate list, keeping only the
// candidate at the specified index.
type filterPos struct {
	index int
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
	attr string
}

func (f *filterAttr) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		if a := c.SelectAttr(f.attr); a != nil {
			p.scratch = append(p.scratch, c)
		}
	}
	p.candidates = p.scratch
}

// filterAttrVal filters the candidate list for elements having
// the specified attribute with the specified value.
type filterAttrVal struct {
	attr, val string
}

func (f *filterAttrVal) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		if a := c.SelectAttr(f.attr); a != nil && a.Value == f.val {
			p.scratch = append(p.scratch, c)
		}
	}
	p.candidates = p.scratch
}

// filterChild filters the candidate list for elements having
// a child element with the specified tag.
type filterChild struct {
	tag string
}

func (f *filterChild) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok && cc.Tag == f.tag {
				p.scratch = append(p.scratch, c)
			}
		}
	}
	p.candidates = p.scratch
}

// filterChildText filters the candidate list for elements having
// a child element with the specified tag and text.
type filterChildText struct {
	tag, text string
}

func (f *filterChildText) apply(p *pather) {
	p.scratch = p.scratch[:0]
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok && cc.Tag == f.tag && cc.Text() == f.text {
				p.scratch = append(p.scratch, c)
			}
		}
	}
	p.candidates = p.scratch
}
