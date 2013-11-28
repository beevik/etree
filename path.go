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
	"strings"
)

const (
	selectSelf                       = iota // "."
	selectParent                            // ".."
	selectChildrenAll                       // "*"
	selectChildrenAllRecursive              // ""
	selectChildrenWithTag                   // "tag"
	selectChildrenContainingTag             // "tag[tag2]"
	selectChildrenContainingTagValue        // "tag[tag2='val']"
	selectChildrenWithAttr                  // "tag[@attr]"
	selectChildrenWithAttrValue             // "tag[@attr='val']"
)

// A selector describes the selection criteria for a path segment.
type selector struct {
	Type  int
	Tag   string
	CTag  string
	CAttr string
	Value string
}

// getSelector converts a segment string into a selector object.
func getSelector(seg string) selector {
	switch seg {
	case "":
		return selector{Type: selectChildrenAllRecursive}
	case ".":
		return selector{Type: selectSelf}
	case "..":
		return selector{Type: selectParent}
	case "*":
		return selector{Type: selectChildrenAll}
	}

	// Find subselector between in [brackets]
	var tag, subselector string
	lindex := strings.IndexByte(seg, '[')
	rindex := strings.IndexByte(seg, ']')
	if lindex > -1 && rindex > -1 && rindex > lindex {
		tag, subselector = seg[0:lindex], seg[lindex+1:rindex]
	}

	// No subselector? Must be a simple child tag selector
	if subselector == "" {
		return selector{Type: selectChildrenWithTag, Tag: seg}
	}

	// Key-value subselector? (e.g., [key='value'])
	eqindex := strings.Index(subselector, "='")
	if eqindex > -1 {
		var key, value string
		rqindex := strings.IndexByte(subselector[eqindex+2:], '\'')
		if rqindex > -1 {
			key, value = subselector[0:eqindex], subselector[eqindex+2:eqindex+2+rqindex]
		}
		if len(key) > 0 && key[0] == '@' {
			return selector{
				Type:  selectChildrenWithAttrValue,
				Tag:   tag,
				CAttr: key[1:],
				Value: value,
			}
		} else {
			return selector{
				Type:  selectChildrenContainingTagValue,
				Tag:   tag,
				CTag:  key,
				Value: value,
			}
		}
	}

	// Must be a simple key selector (e.g., [key])
	if len(subselector) > 0 && subselector[0] == '@' {
		return selector{
			Type:  selectChildrenWithAttr,
			Tag:   tag,
			CAttr: subselector[1:],
		}
	} else {
		return selector{
			Type: selectChildrenContainingTag,
			Tag:  tag,
			CTag: subselector,
		}
	}

}

// A pather is used to traverse an element tree, collecting results
// that match a series of path selectors.
type pather struct {
	results    []*Element        // list of elements matching path query
	stack      []pathNode        // stack for traversing element tree
	candidates []*Element        // scratch array used during traversal
	inResults  map[*Element]bool // tracks which elements are in results
}

// A pathNode represents an element and the remaining path that
// should be applied against it by the pather.
type pathNode struct {
	e    *Element // current element
	path string   // path to apply against current element
}

func newPather() *pather {
	return &pather{
		results:    make([]*Element, 0),
		stack:      make([]pathNode, 0, 1),
		candidates: make([]*Element, 0, 1),
		inResults:  make(map[*Element]bool),
	}
}

func (p *pather) empty() bool {
	return len(p.stack) == 0
}

func (p *pather) push(n pathNode) {
	p.stack = append(p.stack, n)
}

func (p *pather) pop() pathNode {
	n := p.stack[len(p.stack)-1]
	p.stack = p.stack[0 : len(p.stack)-1]
	return n
}

// traverse follows the path from the element e, collecting
// and then returning all elements that match the path's selectors.
func (p *pather) traverse(e *Element, path string) []*Element {
	for p.push(pathNode{e, path}); !p.empty(); {
		p.eval(p.pop())
	}
	return p.results
}

// eval evalutes the current path node by applying the remaining path's
// selector rules against the node's element.
func (p *pather) eval(n pathNode) {
	p.candidates = p.candidates[:0]
	seg, remain := getNextSegment(n.path)

	// Gather all candidate elements that match the current segment
	selector := getSelector(seg)
	switch selector.Type {
	case selectSelf:
		p.candidates = append(p.candidates, n.e)
	case selectParent:
		p.addParent(n.e)
	case selectChildrenAll:
		p.addChildrenAll(n.e)
	case selectChildrenAllRecursive:
		p.addChildrenAllRecursive(n.e)
	case selectChildrenWithTag:
		p.addChildrenWithTag(n.e, &selector)
	case selectChildrenContainingTag:
		p.addChildrenContainingTag(n.e, &selector)
	case selectChildrenContainingTagValue:
		p.addChildrenContainingTagValue(n.e, &selector)
	case selectChildrenWithAttr:
		p.addChildrenWithAttr(n.e, &selector)
	case selectChildrenWithAttrValue:
		p.addChildrenWithAttrValue(n.e, &selector)
	}

	// No path remaining? Then add the candidates to the result set.
	// Otherwise push the candidates on the stack along with the
	// remaining path.
	if remain == "" {
		for _, c := range p.candidates {
			if in := p.inResults[c]; !in {
				p.results = append(p.results, c)
				p.inResults[c] = true
			}
		}
	} else {
		for _, c := range p.candidates {
			p.push(pathNode{c, remain})
		}
	}
}

// getNextSegment splits the path into the next segment and the remaining
// path.
func getNextSegment(path string) (segment, remain string) {
	if i := strings.IndexByte(path, '/'); i > -1 {
		return path[0:i], path[i+1:]
	}
	return path, ""
}

// matchTag returns true if the element's tag matches the selector's
// tag.  A selector tag of "*" matches any element tag.
func matchTag(eTag, sTag string) bool {
	return sTag == "*" || eTag == sTag
}

// addParent adds the selectParent of the element to the candidate list.
func (p *pather) addParent(e *Element) {
	if e.Parent != nil {
		p.candidates = append(p.candidates, e.Parent)
	}
}

// addChildrenAll adds all direct children of the element to the
// candidate list.
func (p *pather) addChildrenAll(e *Element) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok {
			p.candidates = append(p.candidates, c)
		}
	}
}

// addChildrenAllRecursive adds the element to the candidate list.  It then
// performs a depth-first search, adding all subelements below the element
// to the candidate list.
func (p *pather) addChildrenAllRecursive(e *Element) {
	s := elementStack{e}
	for !s.empty() {
		e := s.pop()
		p.candidates = append(p.candidates, e)
		for _, t := range e.Child {
			if ce, ok := t.(*Element); ok {
				s.push(ce)
			}
		}
	}
}

// addChildrenWithTag adds to the candidate list all direct children of
// e with tag selector.Tag.
func (p *pather) addChildrenWithTag(e *Element, s *selector) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && matchTag(c.Tag, s.Tag) {
			p.candidates = append(p.candidates, c)
		}
	}
}

// addChildrenContainingTag adds to the candidate list all direct
// children of e containing a child element with tag selector.Tag
// and a grandchild element with tag selector.CTag.
func (p *pather) addChildrenContainingTag(e *Element, s *selector) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && matchTag(c.Tag, s.Tag) {
			for _, cc := range c.Child {
				if cc, ok := cc.(*Element); ok && matchTag(cc.Tag, s.CTag) {
					p.candidates = append(p.candidates, c)
				}
			}
		}
	}
}

// addChildrenContainingTagValue adds to the candidate list all direct
// children of e containing a child element with tag selector.Tag
// and a grandchild element with tag selector.CTag with a text value
// of s.Value.
func (p *pather) addChildrenContainingTagValue(e *Element, s *selector) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && matchTag(c.Tag, s.Tag) {
			for _, cc := range c.Child {
				if cc, ok := cc.(*Element); ok && matchTag(cc.Tag, s.CTag) && cc.Text() == s.Value {
					p.candidates = append(p.candidates, c)
				}
			}
		}
	}
}

// addChildrenWithAttr adds to the candidate list all direct
// children of e containing a child element with tag selector.Tag
// and an attribute named selector.CAttr.
func (p *pather) addChildrenWithAttr(e *Element, s *selector) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && matchTag(c.Tag, s.Tag) {
			if a := c.SelectAttr(s.CAttr); a != nil {
				p.candidates = append(p.candidates, c)
			}
		}
	}
}

// addChildrenWithAttr adds to the candidate list all direct
// children of e containing a child element with tag selector.Tag
// and an attribute key-value pair with key= selector.CAttr and
// value= selector.Value.
func (p *pather) addChildrenWithAttrValue(e *Element, s *selector) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && matchTag(c.Tag, s.Tag) {
			if a := c.SelectAttr(s.CAttr); a != nil && a.Value == s.Value {
				p.candidates = append(p.candidates, c)
			}
		}
	}
}
