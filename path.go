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

package etree

import (
	"errors"
	"strconv"
	"strings"
)

var (
	errPath = errors.New("etree: invalid path")
)

// A segment is a portion of a path between "/" characters.
// If contains one selector and zero or more [filters].
type segment struct {
	sel     selector
	filters []filter
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

func (seg *segment) apply(e *Element, p *pather) {
	seg.sel.apply(e, p)
	for _, f := range seg.filters {
		f.apply(p)
	}
}

// A pather is used to traverse an element tree, collecting results
// that match a series of path selectors and filters.
type pather struct {
	stack      []node
	results    []*Element
	inResults  map[*Element]bool
	candidates []*Element
}

// A node represents an element and the remaining path segments that
// should be applied against it by the pather.
type node struct {
	e        *Element
	segments []segment
}

func newPather() *pather {
	return &pather{
		stack:      make([]node, 0),
		results:    make([]*Element, 0),
		inResults:  make(map[*Element]bool),
		candidates: make([]*Element, 0),
	}
}

func (p *pather) push(n node) {
	p.stack = append(p.stack, n)
}

func (p *pather) pop() node {
	n := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	return n
}

func (p *pather) empty() bool {
	return len(p.stack) == 0
}

// traverse follows the path from the element e, collecting
// and then returning all elements that match the path's selectors
// and filters.
func (p *pather) traverse(e *Element, pathstr string) []*Element {
	segments := parsePath(pathstr)
	for p.push(node{e, segments}); !p.empty(); {
		p.eval(p.pop())
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
			p.push(node{c, remain})
		}
	}
}

// parsePath parses an XPath-like string describing a path
// through an element tree and returns a slice of segment
// descriptors.
func parsePath(path string) []segment {
	// Path can start with //, but not /
	if strings.HasPrefix(path, "//") {
		path = path[1:]
	} else if strings.HasPrefix(path, "/") {
		panic(errPath)
	}

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
	eqindex := strings.Index(path, "='")
	if eqindex == -1 {
		switch {
		case path[0] == '@':
			return &filterAttr{path[1:]}
		case isInteger(path):
			pos, _ := strconv.Atoi(path)
			if pos == 0 {
				pos = 1
			}
			return &filterPos{pos - 1}
		default:
			return &filterChild{path}
		}
	} else {
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
	return nil
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
				p.candidates = append(p.candidates, c)
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
	result := make([]*Element, 0)
	if f.index < len(p.candidates) {
		result = append(result, p.candidates[f.index])
	}
	p.candidates = result
}

// filterAttr filters the candidate list for elements having
// the specified attribute.
type filterAttr struct {
	attr string
}

func (f *filterAttr) apply(p *pather) {
	result := make([]*Element, 0)
	for _, c := range p.candidates {
		if a := c.SelectAttr(f.attr); a != nil {
			result = append(result, c)
		}
	}
	p.candidates = result
}

// filterAttrVal filters the candidate list for elements having
// the specified attribute with the specified value.
type filterAttrVal struct {
	attr, val string
}

func (f *filterAttrVal) apply(p *pather) {
	result := make([]*Element, 0)
	for _, c := range p.candidates {
		if a := c.SelectAttr(f.attr); a != nil && a.Value == f.val {
			result = append(result, c)
		}
	}
	p.candidates = result
}

// filterChild filters the candidate list for elements having
// a child element with the specified tag.
type filterChild struct {
	tag string
}

func (f *filterChild) apply(p *pather) {
	result := make([]*Element, 0)
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok && cc.Tag == f.tag {
				result = append(result, c)
			}
		}
	}
	p.candidates = result
}

// filterChildText filters the candidate list for elements having
// a child element with the specified tag and text.
type filterChildText struct {
	tag, text string
}

func (f *filterChildText) apply(p *pather) {
	result := make([]*Element, 0)
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok && cc.Tag == f.tag && cc.Text() == f.text {
				result = append(result, c)
			}
		}
	}
	p.candidates = result
}
